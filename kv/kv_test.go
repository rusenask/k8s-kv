package kv

import (
	"testing"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
)

type fakeImplementer struct {
	getcfgMap *v1.ConfigMap

	createdMap *v1.ConfigMap
	updatedMap *v1.ConfigMap

	deletedName    string
	deletedOptions *meta_v1.DeleteOptions
}

func (i *fakeImplementer) Get(name string, options meta_v1.GetOptions) (*v1.ConfigMap, error) {
	return i.getcfgMap, nil
}

func (i *fakeImplementer) Create(cfgMap *v1.ConfigMap) (*v1.ConfigMap, error) {
	i.createdMap = cfgMap
	return i.createdMap, nil
}

func (i *fakeImplementer) Update(cfgMap *v1.ConfigMap) (*v1.ConfigMap, error) {
	i.updatedMap = cfgMap
	return i.updatedMap, nil
}

func (i *fakeImplementer) Delete(name string, options *meta_v1.DeleteOptions) error {
	i.deletedName = name
	i.deletedOptions = options
	return nil
}

func TestGetMap(t *testing.T) {
	fi := &fakeImplementer{
		getcfgMap: &v1.ConfigMap{
			Data: map[string]string{
				"foo": "bar",
			},
		},
	}
	kv, err := New(fi, "app", "b1")
	if err != nil {
		t.Fatalf("failed to get kv: %s", err)
	}

	cfgMap, err := kv.getMap()
	if err != nil {
		t.Fatalf("failed to get map: %s", err)
	}

	if cfgMap.Data["foo"] != "bar" {
		t.Errorf("cfgMap.Data is missing expected key")
	}
}

func TestGet(t *testing.T) {
	fi := &fakeImplementer{
		getcfgMap: &v1.ConfigMap{
			Data: map[string]string{
				"foo": "bar",
			},
		},
	}
	kv, err := New(fi, "app", "b1")
	if err != nil {
		t.Fatalf("failed to get kv: %s", err)
	}

	val, err := kv.Get("foo")
	if err != nil {
		t.Fatalf("failed to get key: %s", err)
	}

	if string(val) != "bar" {
		t.Errorf("expected 'bar' but got: %s", string(val))
	}
}

func TestUpdate(t *testing.T) {
	fi := &fakeImplementer{
		getcfgMap: &v1.ConfigMap{
			Data: map[string]string{
				"a": "a-val",
				"b": "b-val",
				"c": "c-val",
				"d": "d-val",
			},
		},
	}
	kv, err := New(fi, "app", "b1")
	if err != nil {
		t.Fatalf("failed to get kv: %s", err)
	}

	err = kv.Put("b", []byte("updated"))
	if err != nil {
		t.Fatalf("failed to get key: %s", err)
	}

	if fi.updatedMap.Data["b"] != "updated" {
		t.Errorf("b value was not updated")
	}
}
