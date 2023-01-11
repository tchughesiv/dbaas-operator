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

package v1alpha1

import (
	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// notes on writing good spokes https://book.kubebuilder.io/multiversion-tutorial/conversion.html

// ConvertTo converts this DBaaSPlatform to the Hub version (v1beta1).
func (src *DBaaSPlatform) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1beta1.DBaaSPlatform)

	// ObjectMeta
	dst.ObjectMeta = src.ObjectMeta

	// Spec
	dst.Spec = v1beta1.DBaaSPlatformSpec(src.Spec)

	// Status
	dst.Status.Conditions = src.Status.Conditions
	for i := range src.Status.PlatformsStatus {
		dst.Status.PlatformsStatus = append(dst.Status.PlatformsStatus, v1beta1.PlatformStatus{
			PlatformName:   v1beta1.PlatformName(src.Status.PlatformsStatus[i].PlatformName),
			PlatformStatus: v1beta1.PlatformInstlnStatus(src.Status.PlatformsStatus[i].PlatformStatus),
			LastMessage:    src.Status.PlatformsStatus[i].LastMessage,
		})
	}

	return nil
}

// ConvertFrom converts from the Hub version (v1beta1) to this version.
func (dst *DBaaSPlatform) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta1.DBaaSPlatform)

	// ObjectMeta
	dst.ObjectMeta = src.ObjectMeta

	// Spec
	dst.Spec = DBaaSPlatformSpec(src.Spec)

	// Status
	dst.Status.Conditions = src.Status.Conditions
	for i := range src.Status.PlatformsStatus {
		dst.Status.PlatformsStatus = append(dst.Status.PlatformsStatus, PlatformStatus{
			PlatformName:   PlatformsName(src.Status.PlatformsStatus[i].PlatformName),
			PlatformStatus: PlatformsInstlnStatus(src.Status.PlatformsStatus[i].PlatformStatus),
			LastMessage:    src.Status.PlatformsStatus[i].LastMessage,
		})
	}

	return nil
}
