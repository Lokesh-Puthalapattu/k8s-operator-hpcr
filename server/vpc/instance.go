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

package vpc

import (
	"fmt"
	"log"

	"github.com/IBM/vpc-go-sdk/vpcv1"
	"github.com/ibm-hyper-protect/k8s-operator-hpcr/env"
	"github.com/ibm-hyper-protect/k8s-operator-hpcr/vpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	KeyTargetImageName = "TARGET_IMAGE_NAME"
	KeyTargetProfile   = "TARGET_PROFILE"
	KeySubnetID        = "TARGET_SUBNET_ID"
	DefaultProfileName = "bz2e-2x8"
	ServicePrefix      = "k8s-operator-hpcr"
)

type (
	InstanceOptions struct {
		Name        string `json:"name"`
		VpcID       string `json:"vpcID"`
		ProfileName string `json:"profileName"`
		ImageID     string `json:"imageID"`
		ZoneName    string `json:"zone"`
		SubnetID    string `json:"subnetID"`
		UserData    string `json:"userData"`
	}

	CustomResourceSpec struct {
		Contract    string  `json:"contract"`
		SubnetID    *string `json:"subnetID"`
		ProfileName *string `json:"profileName"`
		// specification of the associated config maps
		TargetSelector *metav1.LabelSelector `json:"targetSelector"`
	}

	CustomResource struct {
		metav1.TypeMeta `json:",inline"`
		// Standard object's metadata.
		// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
		// +optional
		metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
		// Specification of the desired behavior of the pod.
		// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
		// +optional
		Spec CustomResourceSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	}

	InstanceConfigResource struct {
		Parent CustomResource `json:"parent"`
	}
)

func CreateVpcInstanceOptions(opt *InstanceOptions) (*vpcv1.CreateInstanceOptions, error) {
	// this is the contract
	options := &vpcv1.CreateInstanceOptions{}
	options.SetInstancePrototype(&vpcv1.InstancePrototypeInstanceByImage{
		// Keys:                    []vpcv1.KeyIdentityIntf{&vpcv1.KeyIdentity{ID: &sshkeyID}},
		Name:              &opt.Name,
		NetworkInterfaces: []vpcv1.NetworkInterfacePrototype{},
		Profile:           &vpcv1.InstanceProfileIdentity{Name: &opt.ProfileName},
		UserData:          &opt.UserData,
		// VolumeAttachments:       []vpcv1.VolumeAttachmentPrototypeInstanceContext{*volumeAttachmentPrototypeModel},
		VPC:                     &vpcv1.VPCIdentity{ID: &opt.VpcID},
		Image:                   &vpcv1.ImageIdentity{ID: &opt.ImageID},
		PrimaryNetworkInterface: &vpcv1.NetworkInterfacePrototype{Subnet: &vpcv1.SubnetIdentity{ID: &opt.SubnetID}},
		Zone:                    &vpcv1.ZoneIdentity{Name: &opt.ZoneName},
	})
	return options, nil
}

func InstanceNameFromUID(uid types.UID) string {
	return fmt.Sprintf("%s-%s", ServicePrefix, uid)
}

func getProfileName(data *InstanceConfigResource, envMap env.Environment) string {
	// check if we have a subnet ID in the config
	if data.Parent.Spec.ProfileName != nil {
		profile := *data.Parent.Spec.ProfileName
		// log this
		log.Printf("Reading profile [%s] from CRD.", profile)
		return profile
	}
	// try to get he profile from the environment
	profile, ok := envMap[KeyTargetProfile]
	if !ok {
		return DefaultProfileName
	}
	// log this
	log.Printf("Reading profile [%s] from environment [%s].", profile, KeyTargetProfile)
	return profile
}

func getImageID(service *vpcv1.VpcV1, envMap env.Environment) (string, error) {
	// try to find the image
	imageName, ok := envMap[KeyTargetImageName]
	if ok {
		log.Printf("Reading image name [%s] from environment [%s].", imageName, KeyTargetImageName)
		// try to find image by name
		return vpc.Findimage(service, imageName)
	}
	// try to find the stock image
	return vpc.FindLatestStockImage(service)
}

func getSubnetID(data *InstanceConfigResource, envMap env.Environment) (string, error) {
	// the ID
	var subnetID string
	// check if we have a subnet ID in the config
	if data.Parent.Spec.SubnetID != nil {
		subnetID = *data.Parent.Spec.SubnetID
		// log this
		log.Printf("Reading Subnet ID [%s] from CRD.", subnetID)
	} else {
		// get the subnet ID from the environment
		subnetIDFromEnv, ok := envMap[KeySubnetID]
		if !ok {
			return "", fmt.Errorf("unable to load the subnet ID from config value [%s]", KeySubnetID)
		}
		// log this
		log.Printf("Reading Subnet ID [%s] from environment [%s].", subnetIDFromEnv, KeySubnetID)
		subnetID = subnetIDFromEnv
	}
	// try to find the subnet
	return subnetID, nil
}

func getSubnet(service *vpcv1.VpcV1, data *InstanceConfigResource, envMap env.Environment) (*vpcv1.Subnet, error) {
	// the ID
	subnetID, err := getSubnetID(data, envMap)
	if err != nil {
		return nil, err
	}
	// try to find the subnet
	subnet, _, err := service.GetSubnet(&vpcv1.GetSubnetOptions{ID: &subnetID})
	return subnet, err
}

func InstanceOptionsFromConfigMap(service *vpcv1.VpcV1, data *InstanceConfigResource, envMap env.Environment) (*InstanceOptions, error) {
	// try to get the subnet
	subnet, err := getSubnet(service, data, envMap)
	if err != nil {
		return nil, err
	}
	// try to get he profile
	profile := getProfileName(data, envMap)
	// try to find the image
	imageID, err := getImageID(service, envMap)
	if err != nil {
		return nil, err
	}
	// convert
	opt := InstanceOptions{
		Name:        InstanceNameFromUID(data.Parent.UID),
		VpcID:       *subnet.VPC.ID,
		ImageID:     imageID,
		ProfileName: profile,
		ZoneName:    *subnet.Zone.Name,
		SubnetID:    *subnet.ID,
		UserData:    data.Parent.Spec.Contract,
	}
	// convert to instance options
	return &opt, nil
}
