terraform {
  required_providers {
    hpcr = {
      source  = "ibm-hyper-protect/hpcr"
      version = ">= 0.3.22"
    }
  }
}

# archive of the folder containing docker-compose file. This folder could create additional resources such as files 
# to be mounted into containers, environment files etc. This is why all of these files get bundled in a tgz file (base64 encoded)
resource "hpcr_tgz" "contract" {
  folder = "compose"
}

locals {
  # contract in clear text
  contract = yamlencode({
    "env" : {
      "type" : "env",
      "logging" : {
        "logDNA" : {
          "ingestionKey" : var.logdna_ingestion_key,
          "hostname" : var.logdna_ingestion_hostname,
        }
      }
    },
    "workload" : {
      "type" : "workload",
      "compose" : {
        "archive" : hpcr_tgz.contract.rendered
      }
    }
  })
}

# In this step we encrypt the fields of the contract and sign the env and workload field. The certificate to execute the 
# encryption it built into the provider and matches the latest HPCR image. If required it can be overridden. 
# We use a temporary, random keypair to execute the signature. This could also be overriden. 
resource "hpcr_contract_encrypted" "contract" {
  contract = local.contract
  cert     = file("../../build/encrypt.crt")
}

locals {
  pod = yamlencode({
    "kind" : "HyperProtectContainerRuntimeOnPrem",
    "apiVersion" : "hpse.ibm.com/v1",
    "metadata" : {
      "name" : "sample-busybox-onprem"
    },
    "spec" : {
      "contract" : hpcr_contract_encrypted.contract.rendered,
      "imageURL" : "http://hpcr-qcow2-image.default:8080/hpcr.qcow2",
      "storagePool" : "images",
      "targetSelector" : {
        "matchLabels" : {
          "app" : "onpremsample"
        }
      }
    }
  })
}

resource "local_file" "contract" {
  content  = hpcr_contract_encrypted.contract.rendered
  filename = "${path.module}/build/contract.yml"
}

resource "local_file" "pod" {
  content  = local.pod
  filename = "${path.module}/build/pod.yml"
}
