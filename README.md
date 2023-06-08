# Cluster Housekeeping

[![Pipeline](https://github.com/carhartl/cluster-housekeeping/actions/workflows/pipeline.yml/badge.svg)](https://github.com/carhartl/cluster-housekeeping/actions/workflows/pipeline.yml)

## Prerequisites

[minikube](https://minikube.sigs.k8s.io/docs/), [Skaffold](https://skaffold.dev/docs/), [Helm](https://helm.sh/docs/):

```bash
brew install minikube skaffold helm chart-testing
```

Optional, for a [Lefthook](https://github.com/evilmartians/lefthook) based Git hooks setup:

```bash
brew install golangci-lint lefthook prettier yamllint && lefthook install
```

## Development

The implementation was build on Kubernetes 1.26.3. Start a [minikube](https://minikube.sigs.k8s.io/docs/) cluster:

```bash
minikube start --kubernetes-version 1.26.3
```

Then start [Skaffold](https://skaffold.dev/docs/) in continuous watch mode:

```bash
skaffold dev
```

## Testing

Test helm chart:

```bash
ct install --chart-dirs . --charts charts/housekeeping
```

## Observe results

```bash
kubectl logs -l service=housekeeping -n housekeeping -f
```

Deploy a non-compliant workload:

```bash
kubectl apply -f - << EOF
apiVersion: v1
kind: Pod
metadata:
  name: bad-nginx
spec:
  containers:
    - image: nginx
      name: nginx
EOF
```

Deploy a compliant workload:

```bash
kubectl apply -f - << EOF
apiVersion: v1
kind: Pod
metadata:
  name: good-nginx
  labels:
    team: test
spec:
  containers:
    - image: bitnami/nginx
      name: nginx
EOF
```

Output (evaluated once per minute):

```
{"pod":"bad-nginx","rule_evaluation":[{"name":"image_prefix","valid":false},{"name":"team_label_present","valid":false},{"name":"recent_start_time","valid":true}],"evaluated_at":"2023-05-25T15:09:45.398228881Z"}
{"pod":"good-nginx","rule_evaluation":[{"name":"image_prefix","valid":true},{"name":"team_label_present","valid":true},{"name":"recent_start_time","valid":true}],"evaluated_at":"2023-05-25T15:09:45.398848381Z"}
```

## Installation

Add Helm repository:

```bash
helm repo add housekeeping https://carhartl.github.io/cluster-housekeeping/
helm repo update
```

To install chart:

```bash
helm install housekeeping housekeeping/housekeeping --namespace housekeeping --create-namespace
```

To uninstall chart:

```bash
helm delete housekeeping
```

## Notes

### Implementation

Assumptions:

- We're not supposed to evaluate pods in the `kube-system` as well as the `housekeeping` namespace, where the service is deployed to.
- We're not supposed to evaluate init containers.

### Alternatives

The [Kyverno](https://kyverno.io/) policy engine allows to write such rules in a declarative way, similar to other Kubernetes resources. Rules can either be executed in audit mode or be strictly enforced via admission controller, with results being available as custom resources to inspect in the cluster (when in audit mode), as well as via Prometheus metrics.

To give an idea, here are the image prefix, team label and recent start time rules implemented as Kyverno policies (audit mode):

```yaml
apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata:
  name: require-bitnami-image
spec:
  validationFailureAction: Audit
  background: true
  rules:
    - name: validate-image
      match:
        any:
          - resources:
              kinds:
                - Pod
      validate:
        message: "Only Bitnami images are allowed."
        pattern:
          spec:
            containers:
              - image: "bitnami/* | docker.io/bitnami/*"
```

```yaml
apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata:
  name: require-team-label
spec:
  validationFailureAction: Audit
  background: true
  rules:
    - name: check-for-label
      match:
        any:
          - resources:
              kinds:
                - Pod
      validate:
        message: "The label `team` is required."
        pattern:
          metadata:
            labels:
              team: "?*"
```

```yaml
apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata:
  name: require-recent-start-time
spec:
  validationFailureAction: Audit
  background: true
  rules:
    - name: check-for-time-of-pod-creation
      match:
        any:
          - resources:
              kinds:
                - Pod
      preconditions:
        any:
          - key: "{{ request.object.status.phase || '' }}"
            operator: Equals
            value: Running
      validate:
        message: "Pods running for more than a 1 week are prohibited."
        deny:
          conditions:
            all:
              - key: "{{ time_since('', '{{request.object.metadata.creationTimestamp}}', '') }}"
                operator: GreaterThan
                value: 168h
```

Not sure how to export the results as json log lines though :)
