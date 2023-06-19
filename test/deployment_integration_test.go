package test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/docker"
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
)

func TestDeployment(t *testing.T) {
	docker.Build(t, "../.", &docker.BuildOptions{
		Tags: []string{"housekeeping:test"},
		Env: map[string]string{
			"DOCKER_TLS_VERIFY":       os.Getenv("DOCKER_TLS_VERIFY"),
			"DOCKER_HOST":             os.Getenv("DOCKER_HOST"),
			"DOCKER_CERT_PATH":        os.Getenv("DOCKER_CERT_PATH"),
			"MINIKUBE_ACTIVE_DOCKERD": os.Getenv("MINIKUBE_ACTIVE_DOCKERD"),
		},
	})

	helmChartPath := "../charts/housekeeping"
	testNamespace := fmt.Sprintf("housekeeping-test-%s", strings.ToLower(random.UniqueId()))
	releaseName := fmt.Sprintf("test-%s", strings.ToLower(random.UniqueId()))

	helmOptions := &helm.Options{
		SetValues: map[string]string{
			"image.repository": "housekeeping",
			"image.tag":        "test",
			"image.pullPolicy": "Never",
		},
		ExtraArgs: map[string][]string{
			"install": []string{"--namespace", testNamespace, "--create-namespace"},
			"delete":  []string{"--namespace", testNamespace},
		},
	}
	// Now that the chart is deployed, verify the deployment.
	// Setup the kubectl config and context. Here we choose to use the defaults, which is:
	// - HOME/.kube/config for the kubectl config file
	// - Current context of the kubectl config file
	kubectlOptions := k8s.NewKubectlOptions("", "", testNamespace)

	helm.Install(t, helmOptions, helmChartPath, releaseName)
	defer helm.Delete(t, helmOptions, releaseName, true)
	defer k8s.DeleteNamespace(t, kubectlOptions, testNamespace)

	retries := 5
	sleep := 2 * time.Second
	k8s.WaitUntilDeploymentAvailable(t, kubectlOptions, "housekeeping", retries, sleep)
}

func TestNamespaceLabels(t *testing.T) {
	docker.Build(t, "../.", &docker.BuildOptions{
		Tags: []string{"housekeeping:test"},
		Env: map[string]string{
			"DOCKER_TLS_VERIFY":       os.Getenv("DOCKER_TLS_VERIFY"),
			"DOCKER_HOST":             os.Getenv("DOCKER_HOST"),
			"DOCKER_CERT_PATH":        os.Getenv("DOCKER_CERT_PATH"),
			"MINIKUBE_ACTIVE_DOCKERD": os.Getenv("MINIKUBE_ACTIVE_DOCKERD"),
		},
	})

	helmChartPath := "../charts/housekeeping"
	testNamespace := fmt.Sprintf("housekeeping-test-%s", strings.ToLower(random.UniqueId()))
	releaseName := fmt.Sprintf("test-%s", strings.ToLower(random.UniqueId()))

	helmOptions := &helm.Options{
		SetValues: map[string]string{
			"image.repository": "housekeeping",
			"image.tag":        "test",
			"image.pullPolicy": "Never",
		},
		ExtraArgs: map[string][]string{
			"install": []string{"--namespace", testNamespace, "--create-namespace"},
			"delete":  []string{"--namespace", testNamespace},
		},
	}
	// Now that the chart is deployed, verify the deployment.
	// Setup the kubectl config and context. Here we choose to use the defaults, which is:
	// - HOME/.kube/config for the kubectl config file
	// - Current context of the kubectl config file
	kubectlOptions := k8s.NewKubectlOptions("", "", testNamespace)

	helm.Install(t, helmOptions, helmChartPath, releaseName)
	defer helm.Delete(t, helmOptions, releaseName, true)
	defer k8s.DeleteNamespace(t, kubectlOptions, testNamespace)

	ns := k8s.GetNamespace(t, kubectlOptions, testNamespace)
	labels := ns.Labels
	var label string
	var ok bool
	if label, ok = labels["pod-security.kubernetes.io/enforce"]; ok {
		if label != "restricted" {
			t.Fatalf("Expected label \"pod-security.kubernetes.io/enforce\" to have value \"restricted\", got %s", label)
		}
	}
	if label, ok = labels["pod-security.kubernetes.io/enforce-version"]; ok {
		if label != "latest" {
			t.Fatalf("Expected label \"pod-security.kubernetes.io/enforce-version\" to have value \"latest\", got %s", label)
		}
	}
}
