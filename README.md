# Terraform Provider Roger

This is a terraform provider for [roger](https://twiki.cern.ch/twiki/bin/view/Main/RogerClient). Roger is more less a functional replacement for the old quattor "sms" tool.

It manages two important pieces of machine state. Whether you want alarms switched on (or, more accurately, whether you want alarms to be masked), and what the current state of the machine is. Optionally you can then use those state transitions to take actions, such as to remove a machine from a load balancer.

For more information about roger see the [documentation](https://twiki.cern.ch/twiki/bin/view/Main/RogerClient)

## Provider usage

To use the provider you just have to declare a provider block:

```terraform
provider "roger" {
  host = "<YOUR-ROGER-SERVER>"
  port = 8201
}

```

To be able to use the Provider valid Kerberos tickets must also be present

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.22

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `make generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

```shell
make testacc
```
