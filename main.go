package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Rule struct {
	Name  string `json:"name"`
	Valid bool   `json:"valid"`
}

type PodEvaluationResult struct {
	Pod            string `json:"pod"`
	RuleEvaluation []Rule `json:"rule_evaluation"`
}

func evaluateImagePrefix(pod corev1.Pod) Rule {
	valid := true
	// NOTE: for now ignoring init containers!
	for _, container := range pod.Spec.Containers {
		if !strings.HasPrefix(container.Image, "bitnami/") {
			valid = false
			break
		}
	}
	return Rule{Name: "image_prefix", Valid: valid}
}

func evaluateTeamLabel(pod corev1.Pod) Rule {
	valid := len(pod.Labels["team"]) > 0
	return Rule{Name: "team_label_present", Valid: valid}
}

func evaluateStartTime(pod corev1.Pod) Rule {
	valid := pod.GetCreationTimestamp().Time.After(time.Now().Add(-7 * 24 * time.Hour))
	return Rule{Name: "recent_start_time", Valid: valid}
}

func evaluatePodCompliance(pod corev1.Pod) PodEvaluationResult {
	return PodEvaluationResult{Pod: pod.Name, RuleEvaluation: []Rule{
		evaluateImagePrefix(pod),
		evaluateTeamLabel(pod),
		evaluateStartTime(pod),
	}}
}

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
		time.Sleep(10 * time.Second)

		pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		for _, pod := range pods.Items {
			if pod.Namespace == "kube-system" || pod.Namespace == "housekeeping" {
				continue
			}
			if pod.Status.Phase == "Running" {
				r := evaluatePodCompliance(pod)
				bytes, _ := json.Marshal(r)
				fmt.Println(string(bytes))
			}
		}
	}
}
