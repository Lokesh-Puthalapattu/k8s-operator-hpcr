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

package networkref

import (
	"log"

	"github.com/ibm-hyper-protect/k8s-operator-hpcr/onprem"
	"github.com/ibm-hyper-protect/k8s-operator-hpcr/server/common"
	C "github.com/ibm-hyper-protect/terraform-provider-hpcr/contract"
	"libvirt.org/go/libvirtxml"
)

// createNetworkRefReadyAction create the action
func createNetworkRefReadyAction(net *libvirtxml.Network) common.Action {

	return func() (*common.ResourceStatus, error) {
		// metadata to attach
		metadata := C.RawMap{
			"Name": net.Name,
		}
		// marshal the network info into metadata
		netStrg, err := onprem.XMLMarshall(net)
		if err == nil {
			metadata["networkXML"] = netStrg
		} else {
			log.Printf("Unable to marshal the network XML, cause: [%v]", err)
		}
		return &common.ResourceStatus{
			Status:      common.Ready,
			Description: netStrg,
			Error:       nil,
			Metadata:    metadata,
		}, nil
	}
}

// CreateSyncAction synchronizes the state of the resource and determines what to do next
func CreateSyncAction(client *onprem.LivirtClient, opt *onprem.NetworkRefOptions) common.Action {
	// checks for the validity of the data disk
	getNetworkRef := onprem.GetNetworkRef(client)
	netXML, err := getNetworkRef(opt)
	if err != nil {
		return common.CreateErrorAction(err)
	}
	// successfully located the network
	return createNetworkRefReadyAction(netXML)
}
