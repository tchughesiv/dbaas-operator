package controllers

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dbaasv1beta1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1beta1"
)

var _ = Describe("DBaaSPlatform controller", func() {
	Describe("trigger reconcile", func() {
		cr := &dbaasv1beta1.DBaaSPlatform{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dbaas-platform",
				Namespace: testNamespace,
			},
		}
		It("should succeed", func() {
			By("checking the DBaaS resource")
			objectKey := client.ObjectKeyFromObject(cr)
			err := dRec.Get(ctx, objectKey, cr)
			Expect(err).NotTo(HaveOccurred())

			Expect(cr.Spec.SyncPeriod).NotTo(BeNil())
			Expect(FindStatusPlatform(cr.Status.PlatformsStatus, "test")).To(BeNil())
			setStatusPlatform(&cr.Status.PlatformsStatus, dbaasv1beta1.PlatformStatus{
				PlatformName:   "test",
				PlatformStatus: dbaasv1beta1.ResultInProgress,
			})
			Expect(FindStatusPlatform(cr.Status.PlatformsStatus, "test")).NotTo(BeNil())
		})
	})
})
