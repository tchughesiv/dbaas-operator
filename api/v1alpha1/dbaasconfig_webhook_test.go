/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var configName = "test-config"
var ns = &v1.Namespace{
	ObjectMeta: metav1.ObjectMeta{
		Name: "test-namespace",
	},
}
var testDBaaSConfig = &DBaaSConfig{
	ObjectMeta: metav1.ObjectMeta{
		Name:      configName,
		Namespace: ns.Name,
	},
}

var _ = Describe("DBaaSConfig Webhook", func() {
	Context("after creating DBaaSConfig", func() {
		BeforeEach(func() {
			By("creating namespace")
			Expect(k8sClient.Create(ctx, ns)).Should(Succeed())

			By("creating DBaaSConfig")
			Expect(k8sClient.Create(ctx, testDBaaSConfig)).Should(Succeed())

			By("checking DBaaSConfig created")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(testDBaaSConfig), &DBaaSConfig{}); err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
		})

		AfterEach(func() {
			By("deleting DBaaSConfig")
			Expect(k8sClient.Delete(ctx, testDBaaSConfig)).Should(Succeed())

			By("checking DBaaSConfig deleted")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(testDBaaSConfig), &DBaaSConfig{})
				if err != nil && errors.IsNotFound(err) {
					return true
				}
				return false
			}, timeout, interval).Should(BeTrue())
		})

		Context("after creating DBaaSConfig of the same inventory namespace", func() {
			It("should not allow creating DBaaSConfig", func() {
				testConfig := &DBaaSConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-config-1",
						Namespace: ns.Name,
					},
					Spec: DBaaSConfigSpec{},
				}
				By("creating DBaaSConfig")
				Expect(k8sClient.Create(ctx, testConfig)).Should(MatchError("admission webhook \"vdbaasconfig.kb.io\" denied the request:" +
					" metadata.Namespace: Invalid value: \"test-namespace\": the namespace test-namespace is already managed by another config"))
			})
		})
	})
})
