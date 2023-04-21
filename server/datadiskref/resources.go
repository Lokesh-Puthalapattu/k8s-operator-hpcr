// Copyright 2023 IBM Corp.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package datasource

package datadiskref

import (
	"github.com/ibm-hyper-protect/k8s-operator-hpcr/onprem"
	"github.com/ibm-hyper-protect/k8s-operator-hpcr/server/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RefDataDiskRefs references data disk refs as related resources
func RefDataDiskRefs(labels *metav1.LabelSelector) common.RelatedResource {
	return common.RefResource(onprem.APIVersion, onprem.ResourceNameDataDiskRefs, labels)
}
