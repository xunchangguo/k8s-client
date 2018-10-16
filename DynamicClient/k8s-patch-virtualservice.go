// Example showing how to patch Kubernetes resources.
package main

import (
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	//  Leave blank for the default context in your kube config.
	context = ""

	//  Name of the VirtualService to weight, and the two weight values.
	virtualServiceName = "service2"
	weight1            = uint32(50)
	weight2            = uint32(50)
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

func setVirtualServiceWeights(client dynamic.Interface, virtualServiceName string, weight1 uint32, weight2 uint32) error {
	//  Create a GVR which represents an Istio Virtual Service.
	virtualServiceGVR := schema.GroupVersionResource{
		Group:    "networking.istio.io",
		Version:  "v1alpha3",
		Resource: "virtualservices",
	}

	//  Weight the two routes - 50/50.
	patchPayload := make([]patchUInt32Value, 2)
	patchPayload[0].Op = "replace"
	patchPayload[0].Path = "/spec/http/0/route/0/weight"
	patchPayload[0].Value = 50
	patchPayload[1].Op = "replace"
	patchPayload[1].Path = "/spec/http/0/route/1/weight"
	patchPayload[1].Value = 50
	patchBytes, _ := json.Marshal(patchPayload)

	//  Apply the patch to the 'service2' service.
	_, err := client.Resource(virtualServiceGVR).Namespace("default").Patch(virtualServiceName, types.JSONPatchType, patchBytes)
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

	// Creates the dynamic interface.
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	//  Re-balance the weights of the hosts in the virtual service.
	setVirtualServiceWeights(dynamicClient, "service2", 50, 50)
}
