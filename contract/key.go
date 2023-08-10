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

package contract

import (
	"fmt"

	E "github.com/IBM/fp-go/either"
	F "github.com/IBM/fp-go/function"
	R "github.com/IBM/fp-go/record"
)

const (
	KEY_TARGET_CONTRACT_PUB_KEY_FILENAME = "TARGET_CONTRACT_PUB_KEY_FILENAME"
)

var (
	lookupCrt = R.Lookup[string, string](KEY_TARGET_CONTRACT_PUB_KEY_FILENAME)
)

// LoadPublicKeyFromEnv locats the contract key from the environment and loads it
var LoadPublicKeyFromEnv = F.Flow3(
	lookupCrt,
	E.FromOption[error, string](func() error {
		return fmt.Errorf("unable to locate the contract certificate from the environment variable [%s].", KEY_TARGET_CONTRACT_PUB_KEY_FILENAME)
	}),
	E.Chain(readFile),
)
