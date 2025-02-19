package operators

import (
	"context"

	g "github.com/onsi/ginkgo"
	o "github.com/onsi/gomega"
	s "github.com/onsi/gomega/gstruct"
	t "github.com/onsi/gomega/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kube-openapi/pkg/util/sets"
	e2e "k8s.io/kubernetes/test/e2e/framework"

	config "github.com/openshift/api/config/v1"
	configclient "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"

	exutil "github.com/openshift/origin/test/extended/util"
)

var _ = g.Describe("[sig-arch] ClusterOperators", func() {
	defer g.GinkgoRecover()

	var clusterOperators []config.ClusterOperator
	whitelistNoNamespace := sets.NewString(
		"cloud-credential",
		"image-registry",
		"machine-api",
		"marketplace",
		"network",
		"operator-lifecycle-manager",
		"operator-lifecycle-manager-catalog",
		"support",
	)
	whitelistNoOperatorConfig := sets.NewString(
		"cloud-credential",
		"cluster-autoscaler",
		"machine-api",
		"machine-config",
		"marketplace",
		"network",
		"operator-lifecycle-manager",
		"operator-lifecycle-manager-catalog",
		"support",
	)

	g.BeforeEach(func() {
		kubeConfig, err := e2e.LoadConfig()
		o.Expect(err).ToNot(o.HaveOccurred())
		configClient, err := configclient.NewForConfig(kubeConfig)
		o.Expect(err).ToNot(o.HaveOccurred())
		clusterOperatorsList, err := configClient.ClusterOperators().List(context.Background(), metav1.ListOptions{})
		o.Expect(err).ToNot(o.HaveOccurred())
		clusterOperators = clusterOperatorsList.Items
	})

	g.Context("should define", func() {
		g.Specify("at least one namespace in their lists of related objects", func() {
			for _, clusterOperator := range clusterOperators {
				if !whitelistNoNamespace.Has(clusterOperator.Name) {
					o.Expect(clusterOperator.Status.RelatedObjects).To(o.ContainElement(isNamespace()), "ClusterOperator: %s", clusterOperator.Name)
				}
			}

		})

		oc := exutil.NewCLI("clusteroperators")
		g.Specify("at least one related object that is not a namespace", func() {
			controlplaneTopology, err := exutil.GetControlPlaneTopology(oc)
			o.Expect(err).NotTo(o.HaveOccurred())

			if *controlplaneTopology == config.ExternalTopologyMode {
				// The packageserver runs in a different cluster along the other controlplane components
				// when the controlplane is external.
				whitelistNoOperatorConfig.Insert("operator-lifecycle-manager-packageserver")
			}
			for _, clusterOperator := range clusterOperators {
				if !whitelistNoOperatorConfig.Has(clusterOperator.Name) {
					o.Expect(clusterOperator.Status.RelatedObjects).To(o.ContainElement(o.Not(isNamespace())), "ClusterOperator: %s", clusterOperator.Name)
				}
			}
		})

	})
})

func isNamespace() t.GomegaMatcher {
	return s.MatchFields(s.IgnoreExtras|s.IgnoreMissing, s.Fields{
		"Resource": o.Equal("namespaces"),
		"Group":    o.Equal(""),
	})
}
