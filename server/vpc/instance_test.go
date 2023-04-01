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
	"encoding/json"
	"os"
	"testing"

	"github.com/ibm-hyper-protect/k8s-operator-hpcr/env"
	"github.com/ibm-hyper-protect/k8s-operator-hpcr/server/common"
	"github.com/ibm-hyper-protect/k8s-operator-hpcr/vpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
)

func loadMap(name string) (map[string]any, error) {
	// load some data
	var data map[string]any
	b, err := os.ReadFile(name)
	if err != nil {
		return data, err
	}
	err = json.Unmarshal(b, &data)
	return data, err
}

func TestOptionsService(t *testing.T) {
	// load some data
	data, err := loadMap("../../samples/create_resource_full.json")
	require.NoError(t, err)

	cfg, err := common.Transcode[InstanceConfigResource](data)
	require.NoError(t, err)

	assert.Equal(t, types.UID("43861249-71b8-490c-ac2a-e7d0028f99e1"), cfg.Parent.UID)
}

func TestInstanceOptionsFromConfigMap(t *testing.T) {
	envMap, err := envFromDotEnv()
	if err != nil {
		t.Skipf("No .env file")
	}

	// load some data
	data, err := loadMap("../samples/create_resource_full.json")
	require.NoError(t, err)

	auth, err := vpc.CreateAuthenticatorFromEnv(envMap)
	require.NoError(t, err)

	service, err := vpc.CreateVpcServiceFromEnv(auth, envMap)
	require.NoError(t, err)

	cfg, err := common.Transcode[*InstanceConfigResource](data)
	require.NoError(t, err)

	io, err := InstanceOptionsFromConfigMap(service, cfg, env.Environment{})
	require.NoError(t, err)

	opt, err := CreateVpcInstanceOptions(io)
	require.NoError(t, err)

	assert.NotNil(t, opt.InstancePrototype)
}
