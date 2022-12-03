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
	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1beta1"
)

var _ = Context("DBaaSInventory Conversion", func() {
	var _ = Describe("Roundtrip", func() {
		Specify("converts to and from the same object", func() {
			pFalse := false
			src := v1alpha1.DBaaSInventory{
				Spec: v1alpha1.DBaaSOperatorInventorySpec{
					DBaaSInventoryPolicy: v1alpha1.DBaaSInventoryPolicy{
						DisableProvisions:    &pFalse,
						ConnectionNamespaces: &[]string{"test", "ha"},
					},
					ProviderRef: v1alpha1.NamespacedName{
						Name:      "trying",
						Namespace: "this",
					},
				},
			}
			intermediate := v1beta1.DBaaSInventory{}
			dst := v1alpha1.DBaaSInventory{}

			Expect(src.ConvertTo(&intermediate)).To(Succeed())
			Expect(dst.ConvertFrom(&intermediate)).To(Succeed())

			Expect(dst).To(Equal(src))
		})
	})
})

var _ = Context("DBaaSConnection Conversion", func() {
	var _ = Describe("Roundtrip", func() {
		Specify("converts to and from the same object", func() {
			instanceID := "testing"
			src := v1alpha1.DBaaSConnection{
				Spec: v1alpha1.DBaaSConnectionSpec{
					InstanceID: instanceID,
				},
			}
			intermediate := v1beta1.DBaaSConnection{}
			dst := v1alpha1.DBaaSConnection{}

			Expect(src.ConvertTo(&intermediate)).To(Succeed())
			Expect(dst.ConvertFrom(&intermediate)).To(Succeed())
			Expect(intermediate.Spec.DatabaseServiceID).To(Equal(instanceID))
			Expect(dst).To(Equal(src))
		})
	})
})
