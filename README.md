# terraform-provider-terrapwner

Terrapwner is a security-focused Terraform provider designed for testing and validating CI/CD pipelines. It provides a set of data sources that enable both red teamers to simulate pipeline abuse scenarios and blue teamers to validate their security controls and exfiltration risks in a controlled manner. The provider offers capabilities to simulate and assess potential security risks through data exfiltration, command execution, and environment probing.

This repository contains:
- A set of security-focused data sources (`internal/provider/`),
- Examples (`examples/`) and generated documentation (`docs/`),
- Miscellaneous meta files.

These files contain the necessary code to create and use the Terrapwner provider. Tutorials for creating Terraform providers can be found on the [HashiCorp Developer](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework) platform.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.23

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `make generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```
