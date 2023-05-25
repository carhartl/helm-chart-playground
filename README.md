# Cluster Housekeeping

[![Pipeline](https://github.com/carhartl/cluster-housekeeping/actions/workflows/pipeline.yml/badge.svg)](https://github.com/carhartl/cluster-housekeeping/actions/workflows/pipeline.yml)

## Prerequisites

[minikube](https://minikube.sigs.k8s.io/docs/), [Skaffold](https://skaffold.dev/docs/), [Helm](https://helm.sh/docs/):

```bash
brew install minikube skaffold helm
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
{"pod":"good-nginx","rule_evaluation":[{"name":"image_prefix","valid":true},{"name":"team_label_present","valid":true},{"name":"recent_start_time","valid":true}]}
{"pod":"bad-nginx","rule_evaluation":[{"name":"image_prefix","valid":false},{"name":"team_label_present","valid":false},{"name":"recent_start_time","valid":true}]}
```

## Notes

### Implementation

Assumptions:

- We're not supposed to evaluate pods in the `kube-system` as well as the `housekeeping` namespace, where the service is deployed to.
- We're not supposed to evaluate init containers.

### Distribution

The implementation is supposed to be distributed as a Helm chart. The chart is available in the `charts` directory, but isn't released (releasable) to a repository yet.

### Alternatives

The [Kyverno](https://kyverno.io/) policy engine allows to write such rules in a declarative way, similar to other Kubernetes resources. Rules can either be executed in audit mode or be strictly enforced via admission controller, with results being available as custom resources to inspect in the cluster (when in audit mode), as well as via Prometheus metrics.

To give an idea, here are the image prefix and team label rules implemented as Kyverno policies (audit mode):

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

Not sure how to export the results as json log lines though :)
