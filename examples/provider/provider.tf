# Copyright (c) HashiCorp, Inc.

terraform {
  required_providers {
    roger = {
      source  = "gitlab.cern.ch/ai-config-team/roger"
      version = "0.1.0"
    }
  }
}

provider "roger" {
  host = "teigicritical-woger-direct.cern.ch"
  port = 8201
}
