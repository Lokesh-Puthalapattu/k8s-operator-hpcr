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
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	C "github.com/ibm-hyper-protect/k8s-operator-hpcr/common"
	"github.com/ibm-hyper-protect/k8s-operator-hpcr/server/common"
	"github.com/ibm-hyper-protect/k8s-operator-hpcr/vpc"
	v1 "k8s.io/api/core/v1"
)

func CreatePingRoute(version, compileTime string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"version": version,
			"compile": compileTime,
		})
	}
}

func syncVPC(req map[string]any) common.Action {

	dbg, err := json.Marshal(req)
	if err == nil {
		log.Printf("Input [%s]", string(dbg))
	}

	env := common.EnvFromConfigMapsOrSecrets(req)

	vpcSvc, err := vpc.CreateVpcServiceFromEnv(env)
	if err != nil {
		return common.CreateErrorAction(err)
	}

	taggingSvc, err := vpc.CreateTaggingServiceFromEnv(env)
	if err != nil {
		return common.CreateErrorAction(err)
	}

	cfg, err := common.Transcode[*InstanceConfigResource](req)
	if err != nil {
		return common.CreateErrorAction(err)
	}

	opt, err := InstanceOptionsFromConfigMap(vpcSvc, cfg, env)
	if err != nil {
		return common.CreateErrorAction(err)
	}

	return CreateSyncAction(vpcSvc, taggingSvc, opt)
}

func finalizeVPC(req map[string]any) common.Action {
	env := common.EnvFromConfigMapsOrSecrets(req)

	service, err := vpc.CreateVpcServiceFromEnv(env)
	if err != nil {
		return common.CreateErrorAction(err)
	}

	cfg, err := common.Transcode[*InstanceConfigResource](req)
	if err != nil {
		return common.CreateErrorAction(err)
	}

	opt, err := InstanceOptionsFromConfigMap(service, cfg, env)
	if err != nil {
		return common.CreateErrorAction(err)
	}

	return CreateFinalizeAction(service, opt)
}

func CreateControllerSyncRoute() gin.HandlerFunc {

	return func(c *gin.Context) {
		log.Printf("synchronizing ...")
		jsonData, err := io.ReadAll(c.Request.Body)
		if err != nil {
			// print stome log
			log.Printf("Error accessing the request body, cause: [%v]", err)
			// Handle error
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		var req map[string]any
		err = json.Unmarshal(jsonData, &req)
		if err != nil {
			// print stome log
			log.Printf("Error during unmarshaling, cause: [%v]", err)
			// Handle error
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		// constuct the action
		action := syncVPC(req)
		// execute and handle
		state, err := action()
		if err != nil {
			// print some log
			log.Printf("Error executing the sync, cause: [%v]", err)
			// Handle error
			c.JSON(http.StatusBadRequest, common.ResourceStatusToResponse(state))
			return
		}
		// done
		resp := common.ResourceStatusToResponse(state)
		// set a retry if we are not ready, yet
		if state.Status != common.Ready {
			resp["resyncAfterSeconds"] = 10
		}
		// done
		c.JSON(http.StatusOK, resp)
	}

}

func CreateControllerFinalizeRoute() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("finalizing ...")

		jsonData, err := io.ReadAll(c.Request.Body)
		if err != nil {
			// Handle error
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		var req map[string]any
		err = json.Unmarshal(jsonData, &req)
		if err != nil {
			// Handle error
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		// constuct the action
		action := finalizeVPC(req)
		// execute and handle
		state, err := action()
		if err != nil {
			// Handle error
			c.JSON(http.StatusOK, common.ResourceStatusToResponse(state))
			return
		}
		// done finalizing
		finalized := state.Status == common.Ready
		resp := gin.H{
			"finalized": finalized,
		}
		if !finalized {
			resp["resyncAfterSeconds"] = 10
		}
		c.JSON(http.StatusOK, resp)
	}
}

func CreateControllerCustomizeRoute() gin.HandlerFunc {
	return func(c *gin.Context) {

		jsonData, err := io.ReadAll(c.Request.Body)
		if err != nil {
			// Handle error
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		// decode the input
		var req map[string]any
		err = json.Unmarshal(jsonData, &req)
		if err != nil {
			// Handle error
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		// transcode to the expected format
		cfg, err := common.Transcode[*InstanceConfigResource](req)
		if err != nil {
			// Handle error
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		// print namespace
		log.Printf("Getting related resources for [%s] in namespace [%s] ...", cfg.Parent.Name, cfg.Parent.Namespace)

		resp := common.CustomizeHookResponse{
			RelatedResourceRules: []*common.RelatedResourceRule{
				{
					ResourceRule: common.ResourceRule{
						APIVersion: C.K8SAPIVersion,
						Resource:   string(v1.ResourceConfigMaps),
					},
					// select by label
					LabelSelector: cfg.Parent.Spec.TargetSelector,
				}, {
					ResourceRule: common.ResourceRule{
						APIVersion: C.K8SAPIVersion,
						Resource:   string(v1.ResourceSecrets),
					},
					// select by label
					LabelSelector: cfg.Parent.Spec.TargetSelector,
				},
			},
		}

		out, _ := json.Marshal(resp)
		log.Printf("CUSTOMIZE response: %s", string(out))

		// done
		c.JSON(http.StatusOK, resp)
	}

}
