# Kubernetes backed KV

Use Kubernetes config maps as key/value store! 

Don't be silly, you can't put a lot of stuff there, when creating a new `KV` max size limit is 1MB because it's the limit
of Etcd value size. Feel free to create as much instances as you want though with different bucket names. 

Usage example:

1. Get minikube or your favourite k8s environment running.

2. In your app you will probably want to use this: https://github.com/kubernetes/client-go/tree/master/examples/in-cluster-client-configuration

3. Get ConfigMaps interface and supply it to this lib:

```
package main

import (
	"fmt"

	"github.com/rusenask/k8s-kv/kv"

	"k8s.io/client-go/kubernetes"
	core_v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

func getImplementer() (implementer core_v1.ConfigMapInterface) {
	cfg, err := clientcmd.BuildConfigFromFlags("", ".kubeconfig") // in your app you could replace it with in-cluster-config
	if err != nil {
		panic(err)
	}

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		panic(err)
	}

	return client.ConfigMaps("default")
}

func main() {
	impl := getImplementer()

	kvdb, err := kv.New(impl, "my-app", "bucket1")
	if err != nil {
		panic(err)
	}

	kvdb.Put("foo", []byte("hello kubernetes world"))

	stored, _ := kvdb.Get("foo")

	fmt.Println(string(stored))
}
```