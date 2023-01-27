package utilkubernetes

import (
	"context"
	"time"

	kapiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	kapiextensions "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	kapiextensionsv1client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
)

// WaitForAllCRDsReady waits for all CRDs to be ready
func WaitForAllCRDReady(ctx context.Context, config *rest.Config) error {
	// TODO: handle connectivity retries (e.g. dial tcp 20.49.158.118:443: connect: connection refused)
	ae, err := kapiextensions.NewForConfig(config)
	if err != nil {
		return err
	}

	crds, err := ae.ApiextensionsV1().CustomResourceDefinitions().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, crd := range crds.Items {
		err := waitForCRDReady(ctx, crd.Name, ae.ApiextensionsV1().CustomResourceDefinitions())
		if err != nil {
			return err
		}
	}

	return nil
}

// waitForCRDReady waits for a CustomResourceDefinition to be ready and registered
func waitForCRDReady(ctx context.Context, name string, cli kapiextensionsv1client.CustomResourceDefinitionInterface) error {
	if err := wait.PollImmediateInfinite(time.Second,
		checkCustomResourceDefinitionIsReady(ctx, cli, name),
	); err != nil {
		return err
	}

	return nil
}

// checkCustomResourceDefinitionIsReady returns a function which polls a
// CustomResourceDefinition and returns its readiness
func checkCustomResourceDefinitionIsReady(ctx context.Context, cli kapiextensionsv1client.CustomResourceDefinitionInterface, name string) func() (bool, error) {
	return func() (bool, error) {
		crd, err := cli.Get(ctx, name, metav1.GetOptions{})
		switch {
		case errors.IsNotFound(err):
			return false, nil
		case err != nil:
			return false, err
		}

		return customResourceDefinitionIsReady(crd), nil
	}
}

// customResourceDefinitionIsReady returns true if a CustomResourceDefinition is
// considered ready
func customResourceDefinitionIsReady(crd *kapiextensionsv1.CustomResourceDefinition) bool {
	for _, cond := range crd.Status.Conditions {
		if cond.Type == kapiextensionsv1.Established &&
			cond.Status == kapiextensionsv1.ConditionTrue {
			return true
		}
	}

	return false
}
