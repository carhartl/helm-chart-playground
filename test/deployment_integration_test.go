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

	options := &helm.Options{
		SetValues: map[string]string{
			"image.repository": "ghcr.io/carhartl/cluster-housekeeping/housekeeping",
			"image.tag":        "latest",
			"image.pullPolicy": "IfNotPresent"},
	}

	releaseName := fmt.Sprintf("test-%s", strings.ToLower(random.UniqueId()))
	helm.Install(t, options, helmChartPath, releaseName)
	defer helm.Delete(t, options, releaseName, true)

	// Now that the chart is deployed, verify the deployment.
	// Setup the kubectl config and context. Here we choose to use the defaults, which is:
	// - HOME/.kube/config for the kubectl config file
	// - Current context of the kubectl config file
	// We also specify that we are working in the default namespace (required to get the Pod)
	kubectlOptions := k8s.NewKubectlOptions("", "", "default")
	retries := 5
	sleep := 2 * time.Second
	k8s.WaitUntilDeploymentAvailable(t, kubectlOptions, "housekeeping", retries, sleep)
}
