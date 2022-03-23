/*
Copyright 2021.

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

package controllers

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("DBaaSConfig controller", func() {
	BeforeEach(assertResourceCreationIfNotExists(&defaultConfig))
	BeforeEach(assertDBaaSResourceStatusUpdated(&defaultConfig, metav1.ConditionTrue, v1alpha1.Ready))

	Describe("reconcile", func() {
		Context("w/ status NotReady", func() {
			config2 := getDefaultConfig(testNamespace)
			config2.Name = "test"
			BeforeEach(assertResourceCreationIfNotExists(&config2))
			BeforeEach(assertDBaaSResourceStatusUpdated(&config2, metav1.ConditionFalse, v1alpha1.DBaaSConfigNotReady))

			It("should return second config with existing config name in status message", func() {
				getConfig := v1alpha1.DBaaSConfig{}
				err := dRec.Get(ctx, client.ObjectKeyFromObject(&config2), &getConfig)
				Expect(err).NotTo(HaveOccurred())
				Expect(getConfig.Status.Conditions).Should(HaveLen(1))
				Expect(getConfig.Status.Conditions[0].Message).Should(Equal(v1alpha1.MsgConfigNotReady + " - " + defaultConfig.GetName()))
			})
		})
	})
})
