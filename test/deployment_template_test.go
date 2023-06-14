package test

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/gruntwork-io/terratest/modules/helm"
)

func TestDeploymentTemplateRendersContainerImage(t *testing.T) {
	helmChartPath := "../charts/housekeeping"

	// Setup the args. For this test, we will set the following input values:
	// - image=nginx:1.15.8
	options := &helm.Options{
		SetValues: map[string]string{"image.repository": "foo", "image.tag": "latest", "image.pullPolicy": "Never"},
	}

	output := helm.RenderTemplate(t, options, helmChartPath, "deployment", []string{"templates/deployment.yaml"})

	var deployment appsv1.Deployment
	helm.UnmarshalK8SYaml(t, output, &deployment)

	podContainers := deployment.Spec.Template.Spec.Containers

	expectedContainerImage := "foo:latest"
	actualContainerImage := podContainers[0].Image
	if actualContainerImage != expectedContainerImage {
		t.Fatalf("Rendered container image (%s) is not expected (%s)", podContainers[0].Image, expectedContainerImage)
	}

	expectedPullPolicy := corev1.PullPolicy("Never")
	actualPullPolicy := podContainers[0].ImagePullPolicy
	if actualPullPolicy != expectedPullPolicy {
		t.Fatalf("Rendered container image pull policy (%s) is not expected (%s)", actualPullPolicy, expectedPullPolicy)
	}
}
