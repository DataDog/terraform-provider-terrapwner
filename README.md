# terraform-provider-terrapwner

Terrapwner is a security-focused Terraform provider designed for testing and validating CI/CD pipelines. It provides a set of data sources that enable both red teamers to simulate pipeline abuse scenarios and blue teamers to validate their security controls and exfiltration risks in a controlled manner. The provider offers capabilities to simulate and assess potential security risks through data exfiltration, command execution, and environment probing.

## Security Notice

⚠️ **IMPORTANT**: This provider is designed for security testing and should be used with caution:
- Only use in controlled environments
- Do not use in production
- Be aware of legal implications
- Follow responsible disclosure practices

## Quick Example

```hcl
terraform {
  required_providers {
    terrapwner = {
      source = "datadog/terrapwner"
    }
  }
}

# Test what commands can be executed
data "terrapwner_local_exec" "test" {
  command = ["whoami"]
}

# Check network connectivity
data "terrapwner_network_probe" "test" {
  type    = "tcp"
  host    = "internal-service"
  port    = 443
  timeout = 5
}
```

## Features

- **Command Execution Testing**: Test what commands can be executed in your CI/CD environment
- **Remote Script Execution**: Test ability to download and execute remote scripts
- **Network Probes**: Check connectivity to internal services, outside world and DNS resolution
- **Data Exfiltration Simulation**: Test data exfiltration capabilities and detection
- **Environment Analysis**: Dump and analyze environment variables and sensitive data
- **State File Analysis**: Retrieve and analyze what sensitive data is stored in Terraform state

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

## Support

For support and questions:
- Check the [documentation](docs/)
- Open an [issue](https://github.com/datadog/terraform-provider-terrapwner/issues)
- Contact the maintainers

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
