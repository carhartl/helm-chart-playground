package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/carhartl/cluster-housekeeping/internal/rules"
)

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	for {
		time.Sleep(60 * time.Second)

		pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		for _, pod := range pods.Items {
			if pod.Namespace == "kube-system" || pod.Namespace == "housekeeping" {
				continue
			}
			if pod.Status.Phase == "Running" {
				r := rules.EvaluatePodCompliance(pod)
				bytes, _ := json.Marshal(r)
				fmt.Println(string(bytes))
			}
		}
	}
}
