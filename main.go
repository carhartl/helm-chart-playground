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

type PodEvaluationResult struct {
	Pod            string           `json:"pod"`
	RuleEvaluation []RuleEvaluation `json:"rule_evaluation"`
}

type RuleEvaluation struct {
	Name  string `json:"name"`
	Valid bool   `json:"valid"`
}

type Validator func(pod corev1.Pod) bool

type Rule struct {
	Name      string
	Validator Validator
}

func (r Rule) Evaluate(pod corev1.Pod) RuleEvaluation {
	return RuleEvaluation{Name: r.Name, Valid: r.Validator(pod)}
}

func NewRule(name string, validator Validator) Rule {
	return Rule{Name: name, Validator: validator}
}

var imagePrefixRule = NewRule("image_prefix",
	func(pod corev1.Pod) bool {
		valid := true
		// NOTE: for now ignoring init containers!
		for _, container := range pod.Spec.Containers {
			if !strings.HasPrefix(strings.Replace(container.Image, "docker.io/", "", 1), "bitnami/") {
				valid = false
				break
			}
		}
		return valid
	})

var teamLabelPresentRule = NewRule("team_label_present",
	func(pod corev1.Pod) bool {
		return len(pod.Labels["team"]) > 0
	})

var recentStartTimeRule = NewRule("recent_start_time",
	func(pod corev1.Pod) bool {
		return pod.GetCreationTimestamp().Time.After(time.Now().Add(-7 * 24 * time.Hour))
	})

func evaluatePodCompliance(pod corev1.Pod) PodEvaluationResult {
	return PodEvaluationResult{Pod: pod.Name, RuleEvaluation: []RuleEvaluation{
		imagePrefixRule.Evaluate(pod),
		teamLabelPresentRule.Evaluate(pod),
		recentStartTimeRule.Evaluate(pod),
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
				r := evaluatePodCompliance(pod)
				bytes, _ := json.Marshal(r)
				fmt.Println(string(bytes))
			}
		}
	}
}
