terraform {
  required_providers {
    roger = {
      source  = "gitlab.cern.ch/ai-config-team/roger"
      version = "0.1.0"
    }
  }
}

provider "roger" {
  host = "<YOUR-ROGER-SERVER>"
  port = 8201
}
