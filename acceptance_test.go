// This test suite is a WIP. Currently just a rudimentary cluster deployment check.
package main

import (
	"log"
	"net/http"
	"os"
	"testing"

	terratestClient "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/assert"
	versionedClient "istio.io/client-go/pkg/clientset/versioned"
	"k8s.io/client-go/tools/clientcmd"
)

func httpOK(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func K8sAcceptance(t *testing.T) {

	t.Parallel()

	kubeconfig := os.Getenv("KUBECONFIG")
	namespace := os.Getenv("NAMESPACE")
	if len(kubeconfig) == 0 || len(namespace) == 0 {
		log.Fatalf("env vars KUBECONFIG and NAMESPACE must be set")
	}

	options := terratestClient.NewKubectlOptions("", kubeconfig, namespace)

	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("failed to created k8s rest client: %s", err)
	}

	ic, err := versionedClient.NewForConfig(restConfig)
	if err != nil {
		log.Fatalf("Failed to create istio client: %s", err)
	}

	vsList, err := ic.NetworkingV1alpha3().VirtualServices(namespace).List(metav1.ListOptions{})
	for _, vs := range vsList.Items {
		// TODO - iterate through each VirtualService and check healthz
	}

	////////////////////////
	// Basic k8s tests
	////////////////////////

	nodes := terratestClient.GetNodes(t, options)
	terratestClient.AreAllNodesReady(t, options)

	////////////////////////
	// Istio Tests
	// Set the port values to your desired configuration.
	// It is important to configure istio's ingress with predefined ports.
	// This is because the ingress gateway uses nodeport, which unless speicified will randomly select a port.
	// Random ports can lead to unintended side-effects, like improper routing when trying to load balance.
	//
	// https://istio.io/latest/docs/reference/config/istio.operator.v1alpha1/#ServicePort
	//
	// TODO - read the IstioOperator config and verify the node ports
	////////////////////////

	options.Namespace = "istio-system"

	ingress := terratestClient.GetService(t, options, "istio-ingressgateway")
	terratestClient.IsServiceAvailable(ingress)

	egress := terratestClient.GetService(t, options, "istio-egressgateway")
	terratestClient.IsServiceAvailable(egress)

	ingressPort, _ := terratestClient.FindNodePortE(ingress, int32(15021))
	assert.Equal(t, ingressPort, int32(31324))

	ingressPort, _ = terratestClient.FindNodePortE(ingress, int32(80))
	assert.Equal(t, ingressPort, int32(31180))

	ingressSslPort, _ := terratestClient.FindNodePortE(ingress, int32(443))
	assert.Equal(t, ingressSslPort, int32(32005))

	ds := terratestClient.GetDaemonSet(t, options, "istio-cni-node")
	assert.Equal(t, int32(len(nodes)), ds.Status.CurrentNumberScheduled)

}
