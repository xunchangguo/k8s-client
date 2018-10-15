// k8s-client project main.go
package main

import (
	"flag"
	"log"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	namespace := flag.String("namespace", "admin", "namespace")
	var apps string
	flag.StringVar(&apps, "apps", "", "applications for restart, with , for split")
	flag.Parse()
	if apps == "" {
		log.Printf("apps is empty, exist.")
	} else {
		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
		// creates the clientset
		client, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}
		apparr := strings.Split(apps, ",")
		scaleMap := make(map[string]int32, len(apparr))
		for _, app := range apparr {
			scale, err := client.ExtensionsV1beta1().Deployments(*namespace).GetScale(app, metav1.GetOptions{})
			if err != nil {
				log.Printf("Failed to get latest version Deployment[%s %s] of scale, %v", *namespace, app, err)
				continue
			}
			scaleMap[app] = scale.Spec.Replicas
			if scale.Spec.Replicas > 0 {
				scale.Spec.Replicas = 0
				_, err = client.ExtensionsV1beta1().Deployments(*namespace).UpdateScale(app, scale)
				if err != nil {
					log.Printf("Failed to scale Deployment[%s %s] to 0, %v", *namespace, app, err)
				} else {
					log.Printf("Success to scale Deployment[%s %s] to 0", *namespace, app)
				}
			} else {
				log.Printf("No need to scale Deployment[%s %s] to 0", *namespace, app)
			}
		}
		time.Sleep(time.Second)
		//start
		for _, app := range apparr {
			scale, err := client.ExtensionsV1beta1().Deployments(*namespace).GetScale(app, metav1.GetOptions{})
			if err != nil {
				log.Printf("Failed to get latest version Deployment[%s %s] of scale, %v", *namespace, app, err)
				continue
			}
			scale.Spec.Replicas = scaleMap[app]
			if scale.Spec.Replicas > 0 {
				_, err = client.ExtensionsV1beta1().Deployments(*namespace).UpdateScale(app, scale)
				if err != nil {
					log.Printf("Failed to scale Deployment[%s %s] to %d, %v", *namespace, app, scale.Spec.Replicas, err)
				} else {
					log.Printf("Success to scale Deployment[%s %s] to %d", *namespace, app, scale.Spec.Replicas)
				}
			} else {
				log.Printf("No need to restore Deployment[%s %s] scale", *namespace, app)
			}
		}
	}
}
