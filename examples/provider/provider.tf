terraform {
  required_providers {
    roger = {
      source  = "barnes-c/roger"
      version = "1.0.14"
    }
  }
}

provider "roger" {
  host = "<YOUR-ROGER-SERVER>"
  port = 8201
}
