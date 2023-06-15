package test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
)

func TestDeployment(t *testing.T) {
	helmChartPath := "../charts/housekeeping"
	testNamespace := fmt.Sprintf("housekeeping-test-%s", strings.ToLower(random.UniqueId()))
	releaseName := fmt.Sprintf("test-%s", strings.ToLower(random.UniqueId()))

	helmOptions := &helm.Options{
		SetValues: map[string]string{
			"image.repository": "ghcr.io/carhartl/cluster-housekeeping/housekeeping",
			"image.tag":        "latest",
			"image.pullPolicy": "IfNotPresent"},
		ExtraArgs: map[string][]string{
			"install":   []string{"--namespace", testNamespace, "--create-namespace"},
			"uninstall": []string{"--namespace", testNamespace},
		},
	}
	helm.Install(t, helmOptions, helmChartPath, releaseName)
	defer helm.Delete(t, helmOptions, releaseName, true)

	// Now that the chart is deployed, verify the deployment.
	// Setup the kubectl config and context. Here we choose to use the defaults, which is:
	// - HOME/.kube/config for the kubectl config file
	// - Current context of the kubectl config file
	kubectlOptions := k8s.NewKubectlOptions("", "", testNamespace)
	retries := 5
	sleep := 2 * time.Second
	k8s.WaitUntilDeploymentAvailable(t, kubectlOptions, "housekeeping", retries, sleep)
}
