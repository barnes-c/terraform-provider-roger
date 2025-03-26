# Copyright (c) HashiCorp, Inc.

resource "state" "my_state" {
  hostname = "myhostname.cern.ch"
  message  = "my message"
  appstate = "production"
}
