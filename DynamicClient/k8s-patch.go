// Example showing how to patch kubernetes resources.
// This is the companion to my article 'Patching Kubernetes Resources in Golang':
//   https://dwmkerr.com/patching-kubernetes-resources-in-golang/
package main

import (
	"encoding/json"
	"fmt"

	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	//  Leave blank for the default context in your kube config.
	context = ""

	//  Name of the replication controller to scale, and the desired number of replicas.
	replicationControllerName = "my-rc"
	replicas                  = uint32(3)
)

//  patchStringValue specifies a patch operation for a string.
type patchStringValue struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

//  patchStringValue specifies a patch operation for a uint32.
type patchUInt32Value struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value uint32 `json:"value"`
}

func scaleReplicationController(clientSet *kubernetes.Clientset, replicasetName string, scale uint32) error {
	payload := []patchUInt32Value{{
		Op:    "replace",
		Path:  "/spec/replicas",
		Value: scale,
	}}
	payloadBytes, _ := json.Marshal(payload)
	_, err := clientSet.
		CoreV1().
		ReplicationControllers("default").
		Patch(replicasetName, types.JSONPatchType, payloadBytes)
	return err
}

func main() {
	//  Get the local kube config.
	fmt.Printf("Connecting to Kubernetes Context %v\n", context)
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{CurrentContext: context}).ClientConfig()
	if err != nil {
		panic(err.Error())
	}

	// Creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	//  Scale our replication controller.
	fmt.Printf("Scaling replication controller %v to %v\n", replicationControllerName, replicas)
	err = scaleReplicationController(clientset, replicationControllerName, replicas)
	if err != nil {
		panic(err.Error())
	}
}
