package rules

import (
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
)

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

func EvaluatePodCompliance(pod corev1.Pod) PodEvaluation {
	return PodEvaluation{Pod: pod.Name, RuleEvaluation: []RuleEvaluation{
		imagePrefixRule.Evaluate(pod),
		teamLabelPresentRule.Evaluate(pod),
		recentStartTimeRule.Evaluate(pod),
	}}
}
