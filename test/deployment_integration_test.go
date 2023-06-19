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
	"github.com/stretchr/testify/suite"
)

var (
	helmChartPath = "../charts/housekeeping"
	testNamespace = fmt.Sprintf("housekeeping-test-%s", strings.ToLower(random.UniqueId()))
	releaseName   = fmt.Sprintf("test-%s", strings.ToLower(random.UniqueId()))

	helmOptions = &helm.Options{
		SetValues: map[string]string{
			"image.repository": "housekeeping",
			"image.tag":        "test",
			"image.pullPolicy": "Never",
		},
		ExtraArgs: map[string][]string{
			"install": {"--namespace", testNamespace, "--create-namespace"},
			"delete":  {"--namespace", testNamespace},
		},
	}

	// Now that the chart is deployed, verify the deployment.
	// Setup the kubectl config and context. Here we choose to use the defaults, which is:
	// - HOME/.kube/config for the kubectl config file
	// - Current context of the kubectl config file
	kubectlOptions = k8s.NewKubectlOptions("", "", testNamespace)
)

type Suite struct {
	suite.Suite
}

func (s *Suite) SetupSuite() {
	docker.Build(s.T(), "../.", &docker.BuildOptions{
		Tags: []string{"housekeeping:test"},
		Env: map[string]string{
			"DOCKER_TLS_VERIFY":       os.Getenv("DOCKER_TLS_VERIFY"),
			"DOCKER_HOST":             os.Getenv("DOCKER_HOST"),
			"DOCKER_CERT_PATH":        os.Getenv("DOCKER_CERT_PATH"),
			"MINIKUBE_ACTIVE_DOCKERD": os.Getenv("MINIKUBE_ACTIVE_DOCKERD"),
		},
	})
	helm.Install(s.T(), helmOptions, helmChartPath, releaseName)
}

func (s *Suite) TearDownSuite() {
	helm.Delete(s.T(), helmOptions, releaseName, true)
	k8s.DeleteNamespace(s.T(), kubectlOptions, testNamespace)
	docker.DeleteImageE(s.T(), "housekeeping:test", nil)
}

func (s *Suite) TestDeployment() {
	retries := 5
	sleep := 2 * time.Second
	k8s.WaitUntilDeploymentAvailable(s.T(), kubectlOptions, "housekeeping", retries, sleep)
}

func (s *Suite) TestNamespaceLabels() {
	ns := k8s.GetNamespace(s.T(), kubectlOptions, testNamespace)
	labels := ns.Labels
	var (
		label string
		ok    bool
	)
	if label, ok = labels["pod-security.kubernetes.io/enforce"]; ok {
		if label != "restricted" {
			s.T().Fatalf("Expected label \"pod-security.kubernetes.io/enforce\" to have value \"restricted\", got %s", label)
		}
	}
	if label, ok = labels["pod-security.kubernetes.io/enforce-version"]; ok {
		if label != "latest" {
			s.T().Fatalf("Expected label \"pod-security.kubernetes.io/enforce-version\" to have value \"latest\", got %s", label)
		}
	}
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(Suite))
}
