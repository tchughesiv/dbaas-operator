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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var testSecret = corev1.Secret{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "test-credentials",
		Namespace: testNamespace,
		Labels: map[string]string{
			"test": "label",
		},
	},
}

var testSecret2 = corev1.Secret{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "test-credentials",
		Namespace: testNamespace,
		Labels: map[string]string{
			"test": "label",
		},
	},
}

var _ = Describe("Create provider object", func() {
	It("should create the expected provider object", func() {
		object := &v1alpha1.DBaaSConnection{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-connection",
				Namespace: "test-namespace",
			},
		}
		result := dRec.createProviderObject(object, "test-kind")

		expected := &unstructured.Unstructured{}
		expected.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   "dbaas.redhat.com",
			Version: "v1alpha1",
			Kind:    "test-kind",
		})
		expected.SetNamespace("test-namespace")
		expected.SetName("test-connection")
		Expect(result).Should(Equal(expected))
	})
})

var _ = Describe("Get DBaaSProvider", func() {
	provider := &v1alpha1.DBaaSProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-provider",
		},
		Spec: v1alpha1.DBaaSProviderSpec{
			Provider: v1alpha1.DatabaseProvider{
				Name: "test-provider",
			},
			InventoryKind:                "testInventoryKind",
			ConnectionKind:               "testConnectionKind",
			InstanceKind:                 "testInstanceKind",
			CredentialFields:             []v1alpha1.CredentialField{},
			AllowsFreeTrial:              false,
			ExternalProvisionURL:         "",
			ExternalProvisionDescription: "",
			InstanceParameterSpecs:       []v1alpha1.InstanceParameterSpec{},
		},
	}
	BeforeEach(assertResourceCreation(provider))
	AfterEach(assertResourceDeletion(provider))

	It("should get the expected DBaaSProvider", func() {
		provider.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("DBaaSProvider"))

		p, err := dRec.getDBaaSProvider("test-provider", ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(p).Should(Equal(provider))
	})
})

var _ = Describe("Get install Namespace", func() {
	It("should get expected install namespace", func() {
		ns, err := GetInstallNamespace()
		Expect(err).NotTo(HaveOccurred())
		Expect(ns).Should(Equal(testNamespace))
	})
})

var _ = Describe("Parse provider object", func() {
	connectionSpec := v1alpha1.DBaaSConnectionSpec{
		InventoryRef: v1alpha1.NamespacedName{
			Name:      "test-inventory",
			Namespace: "test-namespace",
		},
		InstanceID: "test-instance-id",
	}
	uConnection := &unstructured.Unstructured{}
	uConnection.SetUnstructuredContent(make(map[string]interface{}, 1))
	uConnection.UnstructuredContent()["spec"] = connectionSpec
	eConnection := &v1alpha1.DBaaSProviderConnection{
		Spec: connectionSpec,
	}

	inventorySpec := v1alpha1.DBaaSInventorySpec{
		CredentialsRef: &v1alpha1.LocalObjectReference{
			Name: "test-credential-ref",
		},
	}
	uInventory := &unstructured.Unstructured{}
	uInventory.SetUnstructuredContent(make(map[string]interface{}, 1))
	uInventory.UnstructuredContent()["spec"] = inventorySpec
	eInventory := &v1alpha1.DBaaSProviderInventory{
		Spec: inventorySpec,
	}

	DescribeTable("should correctly parse the provider object",
		func(object interface{}, unstructured *unstructured.Unstructured, expected interface{}) {
			err := dRec.parseProviderObject(unstructured, object)
			Expect(err).NotTo(HaveOccurred())
			Expect(object).Should(Equal(expected))
		},
		Entry("parse DBaaSConnection", &v1alpha1.DBaaSProviderConnection{}, uConnection, eConnection),
		Entry("parse DBaaSInventory", &v1alpha1.DBaaSProviderInventory{}, uInventory, eInventory),
	)
})

var _ = Describe("Provider object MutateFn", func() {
	It("should create the expected MutateFn", func() {
		object := &v1alpha1.DBaaSConnection{}
		providerObject := &unstructured.Unstructured{}
		providerObject.SetUnstructuredContent(make(map[string]interface{}, 1))
		connectionSpec := &v1alpha1.DBaaSConnectionSpec{
			InventoryRef: v1alpha1.NamespacedName{
				Name:      "test-inventory",
				Namespace: "test-namespace",
			},
			InstanceID: "test-instance-id",
		}
		fn := dRec.providerObjectMutateFn(object, providerObject, connectionSpec)
		err := fn()
		Expect(err).NotTo(HaveOccurred())

		expected := &unstructured.Unstructured{}
		expected.SetUnstructuredContent(make(map[string]interface{}, 1))
		expected.UnstructuredContent()["spec"] = connectionSpec
		err = ctrl.SetControllerReference(object, expected, dRec.Scheme)
		Expect(err).NotTo(HaveOccurred())

		Expect(providerObject).Should(Equal(expected))
	})
})

var _ = Describe("Watch DBaaS provider Object", func() {
	It("should invoke controller watch with correctly input", func() {
		source := &unstructured.Unstructured{}
		source.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   v1alpha1.GroupVersion.Group,
			Version: v1alpha1.GroupVersion.Version,
			Kind:    "test-kind",
		})
		owner := &v1alpha1.DBaaSInventory{}
		spyController := newSpyController(nil)

		err := dRec.watchDBaaSProviderObject(spyController, owner, "test-kind")
		Expect(err).NotTo(HaveOccurred())
		Eventually(func() bool {
			return spyController.watched(&watchable{
				source: source,
				owner:  owner,
			})
		}, timeout).Should(BeTrue())
	})
})

var _ = Describe("list configs by inventory namespace", func() {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace-new",
		},
	}
	BeforeEach(assertResourceCreationIfNotExists(ns))

	config1 := getDefaultConfig(ns.Name)
	config1.Name = "test-config-1"
	config2 := getDefaultConfig(ns.Name)
	config2.Name = "test-config-2"
	isTrue := true
	config2.Spec.DisableProvisions = &isTrue
	config3 := getDefaultConfig(ns.Name)
	config3.Name = "test-config-3"
	BeforeEach(assertResourceCreationIfNotExists(&config1))
	BeforeEach(assertDBaaSResourceStatusUpdated(&config1, metav1.ConditionTrue, v1alpha1.Ready))
	BeforeEach(assertResourceCreationIfNotExists(&config2))
	BeforeEach(assertDBaaSResourceStatusUpdated(&config2, metav1.ConditionFalse, v1alpha1.DBaaSConfigNotReady))
	BeforeEach(assertResourceCreationIfNotExists(&config3))
	BeforeEach(assertDBaaSResourceStatusUpdated(&config3, metav1.ConditionFalse, v1alpha1.DBaaSConfigNotReady))

	Context("after creating DBaaSConfigs", func() {
		It("should return all the created configs", func() {
			configList, err := dRec.configListByNS(ctx, ns.Name)
			Expect(err).NotTo(HaveOccurred())
			Expect(configList.Items).Should(HaveLen(3))

			for i := range configList.Items {
				if configList.Items[i].Name == config1.Name {
					Expect(apimeta.IsStatusConditionTrue(configList.Items[i].Status.Conditions, v1alpha1.DBaaSConfigReadyType)).Should(BeTrue())
				} else {
					Expect(apimeta.IsStatusConditionTrue(configList.Items[i].Status.Conditions, v1alpha1.DBaaSConfigReadyType)).Should(BeFalse())
				}
			}

			activeConfig := getActiveConfig(configList)
			Expect(activeConfig).Should(Not(BeNil()))
			Expect(activeConfig.Name).Should(Equal(config1.Name))

			rqList := corev1.ResourceQuotaList{}
			Expect(dRec.List(ctx, &rqList, &client.ListOptions{Namespace: ns.Name})).Should(Succeed())
			Expect(rqList.Items).Should(HaveLen(1))
			Expect(rqList.Items[0].Name).Should(Equal("dbaas-" + config1.Name))
			Expect(rqList.Items[0].Spec.Hard).Should(Equal(corev1.ResourceList{
				corev1.ResourceName("count/dbaasconfigs." + v1alpha1.GroupVersion.Group): resource.MustParse("1"),
			}))
			Expect(isOwner(&config1, &rqList.Items[0], dRec.Scheme)).Should(BeTrue())

			inventory := v1alpha1.DBaaSInventory{ObjectMeta: metav1.ObjectMeta{Namespace: ns.Name}}
			Expect(canProvision(inventory, activeConfig)).Should(BeTrue())
			activeConfig.Spec.DisableProvisions = &isTrue
			Expect(canProvision(inventory, activeConfig)).Should(BeFalse())

			// override config setting
			isFalse := false
			inventory.Spec.DisableProvisions = &isFalse
			Expect(canProvision(inventory, activeConfig)).Should(BeTrue())

			// check nil config
			Expect(canProvision(inventory, nil)).Should(BeFalse())
		})

		It("should, upon deletion, make another config active", func() {
			Expect(dRec.Delete(ctx, &config1)).Should(Succeed())
			Expect(dRec.Delete(ctx, &config3)).Should(Succeed())
			By("checking the resources deleted")
			Eventually(func() bool {
				err := dRec.Get(ctx, client.ObjectKeyFromObject(&config1), &config1)
				if err != nil && errors.IsNotFound(err) {
					return true
				}
				return false
			}, timeout).Should(BeTrue())
			Eventually(func() bool {
				err := dRec.Get(ctx, client.ObjectKeyFromObject(&config3), &config3)
				if err != nil && errors.IsNotFound(err) {
					return true
				}
				return false
			}, timeout).Should(BeTrue())

			By("checking the DBaaS resource status")
			Eventually(func() (bool, error) {
				err := dRec.Get(ctx, client.ObjectKeyFromObject(&config2), &config2)
				if err != nil {
					return false, err
				}
				return apimeta.IsStatusConditionTrue(config2.Status.Conditions, v1alpha1.DBaaSConfigReadyType), nil
			}, timeout).Should(BeTrue())

			configList, err := dRec.configListByNS(ctx, ns.Name)
			Expect(err).NotTo(HaveOccurred())
			Expect(configList.Items).Should(HaveLen(1))

			activeConfig := getActiveConfig(configList)
			Expect(activeConfig).Should(Not(BeNil()))
			Expect(activeConfig.Name).Should(Equal(config2.Name))

			inventory := v1alpha1.DBaaSInventory{ObjectMeta: metav1.ObjectMeta{Namespace: ns.Name}}
			Expect(canProvision(inventory, activeConfig)).Should(BeFalse())
		})
	})
})

var _ = Describe("Check inventory", func() {
	BeforeEach(assertResourceCreationIfNotExists(&testSecret))
	BeforeEach(assertResourceCreationIfNotExists(&testSecret2))
	BeforeEach(assertResourceCreationIfNotExists(mongoProvider))
	BeforeEach(assertResourceCreationIfNotExists(crunchyProvider))
	BeforeEach(assertResourceCreationIfNotExists(&defaultConfig))
	BeforeEach(assertDBaaSResourceStatusUpdated(&defaultConfig, metav1.ConditionTrue, v1alpha1.Ready))

	Context("after creating DBaaSInventory", func() {
		inventoryName := "test-check-inventory"
		createdDBaaSInventory := &v1alpha1.DBaaSInventory{
			ObjectMeta: metav1.ObjectMeta{
				Name:      inventoryName,
				Namespace: testNamespace,
			},
			Spec: v1alpha1.DBaaSOperatorInventorySpec{
				ProviderRef: v1alpha1.NamespacedName{
					Name: testProviderName,
				},
				DBaaSInventorySpec: v1alpha1.DBaaSInventorySpec{
					CredentialsRef: &v1alpha1.LocalObjectReference{
						Name: testSecret.Name,
					},
				},
			},
		}
		createdDBaaSInventory2 := &v1alpha1.DBaaSInventory{
			ObjectMeta: metav1.ObjectMeta{
				Name:      inventoryName + "-2",
				Namespace: testNamespace,
			},
			Spec: v1alpha1.DBaaSOperatorInventorySpec{
				ProviderRef: v1alpha1.NamespacedName{
					Name: "crunchy-bridge-registration",
				},
				DBaaSInventorySpec: v1alpha1.DBaaSInventorySpec{
					CredentialsRef: &v1alpha1.LocalObjectReference{
						Name: testSecret2.Name,
					},
				},
			},
		}
		lastTransitionTime := getLastTransitionTimeForTest()
		providerInventoryStatus := &v1alpha1.DBaaSInventoryStatus{
			Instances: []v1alpha1.Instance{
				{
					InstanceID: "testInstanceID",
					Name:       "testInstance",
					InstanceInfo: map[string]string{
						"testInstanceInfo": "testInstanceInfo",
					},
				},
			},
			Conditions: []metav1.Condition{
				{
					Type:               "SpecSynced",
					Status:             metav1.ConditionTrue,
					Reason:             "SyncOK",
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
				},
			},
		}
		BeforeEach(assertInventoryCreationWithProviderStatus(createdDBaaSInventory, metav1.ConditionTrue, testInventoryKind, providerInventoryStatus))
		BeforeEach(assertInventoryCreationWithProviderStatus(createdDBaaSInventory2, metav1.ConditionTrue, "CrunchyBridgeInventory", providerInventoryStatus))
		AfterEach(assertResourceDeletion(createdDBaaSInventory))
		AfterEach(assertResourceDeletion(createdDBaaSInventory2))

		Context("after creating DBaaSConnection", func() {
			connectionName := "test-check-inventory-connection"
			instanceID := "test-instanceID"
			DBaaSConnectionSpec := &v1alpha1.DBaaSConnectionSpec{
				InventoryRef: v1alpha1.NamespacedName{
					Name:      inventoryName,
					Namespace: testNamespace,
				},
				InstanceID: instanceID,
			}
			createdDBaaSConnection := &v1alpha1.DBaaSConnection{
				ObjectMeta: metav1.ObjectMeta{
					Name:      connectionName,
					Namespace: testNamespace,
				},
				Spec: *DBaaSConnectionSpec,
			}
			BeforeEach(assertResourceCreation(createdDBaaSConnection))
			AfterEach(assertResourceDeletion(createdDBaaSConnection))

			When("check the right inventory", func() {
				It("should return the inventory without error", func() {
					i, validNS, provision, err := dRec.checkInventory(v1alpha1.NamespacedName{
						Name:      inventoryName,
						Namespace: testNamespace,
					}, createdDBaaSConnection, func(reason string, message string) {
						cond := metav1.Condition{
							Type:    v1alpha1.DBaaSConnectionReadyType,
							Status:  metav1.ConditionFalse,
							Reason:  reason,
							Message: message,
						}
						apimeta.SetStatusCondition(&createdDBaaSConnection.Status.Conditions, cond)
					}, ctx, ctrl.LoggerFrom(ctx))

					Expect(err).NotTo(HaveOccurred())
					Expect(validNS).To(Equal(true))
					Expect(provision).To(Equal(true))
					Expect(i.Name).Should(Equal(createdDBaaSInventory.Name))
					Expect(i.Spec).Should(Equal(createdDBaaSInventory.Spec))

					getSecret := corev1.Secret{}
					err = dRec.Get(ctx, client.ObjectKeyFromObject(&testSecret), &getSecret)
					Expect(err).NotTo(HaveOccurred())
					labels := getSecret.GetLabels()
					Expect(labels).Should(Not(BeNil()))
					Expect(labels["test"]).Should(Equal("label"))
					Expect(labels[v1alpha1.TypeLabelKeyMongo]).Should(Equal(v1alpha1.TypeLabelValue))

					getSecret2 := corev1.Secret{}
					err = dRec.Get(ctx, client.ObjectKeyFromObject(&testSecret2), &getSecret2)
					Expect(err).NotTo(HaveOccurred())
					labels2 := getSecret2.GetLabels()
					Expect(labels2).Should(Not(BeNil()))
					Expect(labels2["test"]).Should(Equal("label"))
					Expect(labels2[v1alpha1.TypeLabelKey]).Should(Equal(v1alpha1.TypeLabelValue))
				})
			})

			When("check an inventory not exists", func() {
				It("should return error", func() {
					_, _, _, err := dRec.checkInventory(v1alpha1.NamespacedName{
						Name:      "test-check-not-exist-inventory",
						Namespace: testNamespace,
					}, createdDBaaSConnection, func(reason string, message string) {
						cond := metav1.Condition{
							Type:    v1alpha1.DBaaSConnectionReadyType,
							Status:  metav1.ConditionFalse,
							Reason:  reason,
							Message: message,
						}
						apimeta.SetStatusCondition(&createdDBaaSConnection.Status.Conditions, cond)
					}, ctx, ctrl.LoggerFrom(ctx))

					Expect(err).To(HaveOccurred())
					assertConnectionDBaaSStatus(createdDBaaSConnection.Name, createdDBaaSConnection.Namespace, metav1.ConditionFalse)
				})
			})
		})
	})

	Context("after creating not ready DBaaSInventory", func() {
		inventoryName := "test-check-inventory-not-ready"
		createdDBaaSInventory := &v1alpha1.DBaaSInventory{
			ObjectMeta: metav1.ObjectMeta{
				Name:      inventoryName,
				Namespace: testNamespace,
			},
			Spec: v1alpha1.DBaaSOperatorInventorySpec{
				ProviderRef: v1alpha1.NamespacedName{
					Name: testProviderName,
				},
				DBaaSInventorySpec: v1alpha1.DBaaSInventorySpec{
					CredentialsRef: &v1alpha1.LocalObjectReference{
						Name: testSecret.Name,
					},
				},
			},
		}
		BeforeEach(assertResourceCreation(createdDBaaSInventory))
		AfterEach(assertResourceDeletion(createdDBaaSInventory))

		Context("after creating DBaaSConnection", func() {
			connectionName := "test-check-not-ready-inventory-connection"
			instanceID := "test-instanceID"
			DBaaSConnectionSpec := &v1alpha1.DBaaSConnectionSpec{
				InventoryRef: v1alpha1.NamespacedName{
					Name:      inventoryName,
					Namespace: testNamespace,
				},
				InstanceID: instanceID,
			}
			createdDBaaSConnection := &v1alpha1.DBaaSConnection{
				ObjectMeta: metav1.ObjectMeta{
					Name:      connectionName,
					Namespace: testNamespace,
				},
				Spec: *DBaaSConnectionSpec,
			}
			BeforeEach(assertResourceCreation(createdDBaaSConnection))
			AfterEach(assertResourceDeletion(createdDBaaSConnection))

			When("check an not ready inventory", func() {
				It("should return error", func() {
					_, _, _, err := dRec.checkInventory(v1alpha1.NamespacedName{
						Name:      inventoryName,
						Namespace: testNamespace,
					}, createdDBaaSConnection, func(reason string, message string) {
						cond := metav1.Condition{
							Type:    v1alpha1.DBaaSConnectionReadyType,
							Status:  metav1.ConditionFalse,
							Reason:  reason,
							Message: message,
						}
						apimeta.SetStatusCondition(&createdDBaaSConnection.Status.Conditions, cond)
					}, ctx, ctrl.LoggerFrom(ctx))

					Expect(err).To(HaveOccurred())
					assertConnectionDBaaSStatus(createdDBaaSConnection.Name, createdDBaaSConnection.Namespace, metav1.ConditionFalse)
				})
			})
		})
	})
})

var _ = Describe("Reconcile Provider Resource", func() {
	BeforeEach(assertResourceCreationIfNotExists(&testSecret))
	BeforeEach(assertResourceCreationIfNotExists(mongoProvider))
	BeforeEach(assertResourceCreationIfNotExists(&defaultConfig))
	BeforeEach(assertDBaaSResourceStatusUpdated(&defaultConfig, metav1.ConditionTrue, v1alpha1.Ready))

	Context("after creating DBaaSInventory", func() {
		inventoryName := "test-reconcile-provider-resource-inventory"
		createdDBaaSInventory := &v1alpha1.DBaaSInventory{
			ObjectMeta: metav1.ObjectMeta{
				Name:      inventoryName,
				Namespace: testNamespace,
			},
			Spec: v1alpha1.DBaaSOperatorInventorySpec{
				ProviderRef: v1alpha1.NamespacedName{
					Name: testProviderName,
				},
				DBaaSInventorySpec: v1alpha1.DBaaSInventorySpec{
					CredentialsRef: &v1alpha1.LocalObjectReference{
						Name: testSecret.Name,
					},
				},
			},
		}
		BeforeEach(assertResourceCreation(createdDBaaSInventory))
		AfterEach(assertResourceDeletion(createdDBaaSInventory))

		When("reconcile provider resource with invalid provider", func() {
			It("should return error", func() {
				createdDBaaSInventory.Spec.ProviderRef.Name = "test-reconcile-provider-resource-invalid-provider"
				_, err := dRec.reconcileProviderResource(createdDBaaSInventory.Spec.ProviderRef.Name,
					createdDBaaSInventory,
					func(provider *v1alpha1.DBaaSProvider) string {
						return provider.Spec.InventoryKind
					},
					func() interface{} {
						return createdDBaaSInventory.Spec.DeepCopy()
					},
					func() interface{} {
						return &v1alpha1.DBaaSProviderInventory{}
					},
					func(i interface{}) metav1.Condition {
						providerInventory := i.(*v1alpha1.DBaaSProviderInventory)
						return mergeInventoryStatus(createdDBaaSInventory, providerInventory)
					},
					func() *[]metav1.Condition {
						return &createdDBaaSInventory.Status.Conditions
					},
					v1alpha1.DBaaSInventoryReadyType,
					ctx,
					ctrl.LoggerFrom(ctx),
				)

				Expect(err).To(HaveOccurred())
				assertInventoryDBaaSStatus(createdDBaaSInventory.Name, createdDBaaSInventory.Namespace, metav1.ConditionFalse)
			})
		})
	})
})

func getLastTransitionTimeForTest() time.Time {
	lastTransitionTime, err := time.Parse(time.RFC3339, "2021-06-30T22:17:55-04:00")
	Expect(err).NotTo(HaveOccurred())
	return lastTransitionTime.In(time.Local)
}
