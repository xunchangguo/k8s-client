// test project main.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func main() {
	//	namespace := flag.String("namespace", "admin", "namespace")
	var apps string
	flag.StringVar(&apps, "apps", "", "applications for restart, with , for split")

	kubeconfig := flag.String("kubeconfig", "", "Path to a kube config. Only required if out-of-cluster.")
	username := flag.String("username", "", "k8s username")
	password := flag.String("password", "", "k8s password")

	flag.Parse()
	if apps == "" {
		log.Printf("apps is empty, exist.")
	} else {
		var config *rest.Config
		var err error
		if *kubeconfig != "" {
			config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
		} else if *username != "" {
			apiconfig := clientcmdapi.NewConfig()
			apiconfig.Clusters["default"] = &clientcmdapi.Cluster{
				Server:                "https://172.17.81.55:6443",
				InsecureSkipTLSVerify: true,
			}
			apiconfig.AuthInfos["default"] = &clientcmdapi.AuthInfo{
				Username: *username,
				Password: *password,
			}
			apiconfig.Contexts["default"] = &clientcmdapi.Context{
				Cluster:  "default",
				AuthInfo: "default",
			}
			apiconfig.CurrentContext = "default"
			clientBuilder := clientcmd.NewNonInteractiveClientConfig(*apiconfig, "default", &clientcmd.ConfigOverrides{}, nil)
			config, err = clientBuilder.ClientConfig()
		} else {
			config, err = rest.InClusterConfig()
		}
		//config, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
		//install schema，非常重要
		v1beta1.AddToScheme(scheme.Scheme)

		// creates the clientset
		client, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}
		data, err := client.RESTClient().Get().AbsPath("apis/metrics.k8s.io/v1beta1/nodes").DoRaw()
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("%s", string(data))
		//TODO 这种方式有问题，元数据取不到,解决方法：install, v1beta1.AddToScheme(scheme.Scheme)
		mlist := v1beta1.NodeMetricsList{}
		err = json.Unmarshal(data, &mlist)
		if err != nil {
			panic(err.Error())
		}
		for _, m := range mlist.Items {
			fmt.Printf("%v\n", m.GetObjectMeta())
		}
		fmt.Println("----------------")
		//TODO 这种方式也有问题，元数据取不到,问题和上面一样，解决方法：install, v1beta1.AddToScheme(scheme.Scheme)
		err = client.RESTClient().Get().AbsPath("apis/metrics.k8s.io/v1beta1/nodes").Do().Into(&mlist)
		if err != nil {
			panic(err.Error())
		}
		for _, m := range mlist.Items {
			fmt.Printf("%v\n", m.GetName())
		}
		fmt.Println("----------------")
		//这种方式是取到数据后自己使用Codes decode，一样要先 install schemea
		obj, _, err := scheme.Codecs.UniversalDeserializer().Decode(data, nil, nil)
		if err != nil {
			panic(err.Error())
		}
		nlist := obj.(*v1beta1.NodeMetricsList)
		for _, m := range nlist.Items {
			fmt.Printf("%v\n", m.GetName())
		}
		fmt.Println("----------------")

		//use dynamic client, 个人感觉不是太方便
		gv := &schema.GroupVersion{Group: "metrics.k8s.io", Version: "v1beta1"}
		config.ContentConfig = rest.ContentConfig{GroupVersion: gv}
		config.APIPath = "/apis"

		dynamicClient, _ := dynamic.NewClient(config)
		resource := &metav1.APIResource{
			Name:       "nodes",
			Namespaced: false,
		}
		got, err := dynamicClient.ParameterCodec(scheme.ParameterCodec).Resource(resource, "").List(metav1.ListOptions{})
		if err != nil {
			fmt.Printf("%v", err)
		}
		fmt.Printf("%v \n", got.(*unstructured.UnstructuredList))

	}
}
