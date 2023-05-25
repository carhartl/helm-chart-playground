package rules

import (
	"time"

	corev1 "k8s.io/api/core/v1"
)

type PodEvaluation struct {
	Pod            string           `json:"pod"`
	RuleEvaluation []RuleEvaluation `json:"rule_evaluation"`
	EvaluatedAt    time.Time        `json:"evaluated_at"`
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
