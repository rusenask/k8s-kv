package kv

import (
	"errors"
	"strings"
	"sync"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
)

// errors
var (
	ErrNotFound = errors.New("not found")
)

type KV struct {
	app         string
	bucket      string
	implementer ConfigMapInterface
	mu          *sync.RWMutex
}

type ConfigMapInterface interface {
	Get(name string, options meta_v1.GetOptions) (*v1.ConfigMap, error)
	Create(cfgMap *v1.ConfigMap) (*v1.ConfigMap, error)
	Update(cfgMap *v1.ConfigMap) (*v1.ConfigMap, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
}

func New(implementer ConfigMapInterface, app, bucket string) (*KV, error) {
	kv := &KV{
		implementer: implementer,
		app:         app,
		bucket:      bucket,
		mu:          &sync.RWMutex{},
	}

	_, err := kv.getMap()
	if err != nil {
		return nil, err
	}

	return kv, nil

}

func (k *KV) Teardown() error {
	return k.implementer.Delete(k.bucket, &meta_v1.DeleteOptions{})
}

func (k *KV) getMap() (*v1.ConfigMap, error) {
	cfgMap, err := k.implementer.Get(k.bucket, meta_v1.GetOptions{})
	if err != nil {
		// creating
		if apierrors.IsNotFound(err) {
			return k.newConfigMapsObject()
		}
		return nil, err
	}

	if cfgMap.Data == nil {
		cfgMap.Data = make(map[string]string)
	}

	// it's there, nothing to do
	return cfgMap, nil
}

func (k *KV) newConfigMapsObject() (*v1.ConfigMap, error) {

	var lbs labels

	lbs.init()

	// apply labels
	lbs.set("BUCKET", k.bucket)
	lbs.set("APP", k.app)
	lbs.set("OWNER", "K8S-KV")

	// create and return configmap object
	cfgMap := &v1.ConfigMap{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:   k.bucket,
			Labels: lbs.toMap(),
		},
		Data: map[string]string{},
	}

	cm, err := k.implementer.Create(cfgMap)
	if err != nil {
		return nil, err
	}

	return cm, nil
}

func (k *KV) saveMap(cfgMap *v1.ConfigMap) error {
	_, err := k.implementer.Update(cfgMap)
	return err
}

func (k *KV) Put(key string, value []byte) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	m, err := k.getMap()
	if err != nil {
		return err
	}

	m.Data[key] = string(value)

	return k.saveMap(m)
}

func (k *KV) Get(key string) (value []byte, err error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	m, err := k.getMap()
	if err != nil {
		return nil, err
	}

	val, ok := m.Data[key]
	if !ok {
		return []byte(""), ErrNotFound
	}

	return []byte(val), nil
}

func (k *KV) Delete(key string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	m, err := k.getMap()
	if err != nil {
		return err
	}

	delete(m.Data, key)

	return k.saveMap(m)
}

func (k *KV) List(prefix string) (data map[string][]byte, err error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	m, err := k.getMap()
	if err != nil {
		return
	}

	data = make(map[string][]byte)
	for key, val := range m.Data {
		if strings.HasPrefix(key, prefix) {
			data[key] = []byte(val)
		}
	}
	return
}

// labels is a map of key value pairs to be included as metadata in a configmap object.
type labels map[string]string

func (lbs *labels) init()                { *lbs = labels(make(map[string]string)) }
func (lbs labels) get(key string) string { return lbs[key] }
func (lbs labels) set(key, val string)   { lbs[key] = val }

func (lbs labels) keys() (ls []string) {
	for key := range lbs {
		ls = append(ls, key)
	}
	return
}

func (lbs labels) match(set labels) bool {
	for _, key := range set.keys() {
		if lbs.get(key) != set.get(key) {
			return false
		}
	}
	return true
}

func (lbs labels) toMap() map[string]string { return lbs }

func (lbs *labels) fromMap(kvs map[string]string) {
	for k, v := range kvs {
		lbs.set(k, v)
	}
}
