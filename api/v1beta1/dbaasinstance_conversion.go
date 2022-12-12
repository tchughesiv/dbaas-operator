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

package v1beta1

import (
	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts this DBaaSInstance to the Hub version (v1alpha1).
func (src *DBaaSInstance) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha1.DBaaSInstance)

	// ObjectMeta
	dst.ObjectMeta = src.ObjectMeta

	// Spec
	dst.Spec.CloudProvider = src.Spec.CloudProvider
	dst.Spec.CloudRegion = src.Spec.CloudRegion
	dst.Spec.InventoryRef = v1alpha1.NamespacedName(src.Spec.InventoryRef)
	dst.Spec.Name = src.Spec.Name
	dst.Spec.OtherInstanceParams = src.Spec.OtherInstanceParams

	// Status
	dst.Status.Conditions = src.Status.Conditions
	dst.Status.InstanceID = src.Status.InstanceID
	dst.Status.InstanceInfo = src.Status.InstanceInfo
	dst.Status.Phase = v1alpha1.DBaasInstancePhase(src.Status.Phase)

	return nil
}

// ConvertFrom converts from the Hub version (v1alpha1) to this version.
func (dst *DBaaSInstance) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha1.DBaaSInstance)

	// ObjectMeta
	dst.ObjectMeta = src.ObjectMeta

	// Spec
	dst.Spec.CloudProvider = src.Spec.CloudProvider
	dst.Spec.CloudRegion = src.Spec.CloudRegion
	dst.Spec.InventoryRef = NamespacedName(src.Spec.InventoryRef)
	dst.Spec.Name = src.Spec.Name
	dst.Spec.OtherInstanceParams = src.Spec.OtherInstanceParams

	// Status
	dst.Status.Conditions = src.Status.Conditions
	dst.Status.InstanceID = src.Status.InstanceID
	dst.Status.InstanceInfo = src.Status.InstanceInfo
	dst.Status.Phase = DBaasInstancePhase(src.Status.Phase)

	return nil
}
