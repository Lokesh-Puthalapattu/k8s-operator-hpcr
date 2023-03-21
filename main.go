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

package main

import (
	"log"
	"os"

	"github.com/ibm-hyper-protect/k8s-operator-hpcr/cli"
)

// version number, will be injected by the build
var (
	version  string
	compiled string
	commit   string
)

func main() {
	err := cli.CreateApp(version, compiled, commit).Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
