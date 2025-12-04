# Terraform Provider The Bastion

This provider can be used to managed various resource on [The Bastion](https://github.com/ovh/the-bastion)

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.10
- [Go](https://golang.org/doc/install) >= 1.24

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Using the provider

The provider documentation can be found [here](https://registry.terraform.io/providers/adfinis/bastion).

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `make generate`.

In order to run the full suite of Acceptance tests, run `make testacc-up && make testacc`.


```shell
make testacc-up
make testacc
```

## License

GPL-3.0-or-later

## Author

Created by [Adfinis AG](https://adfinis.com/) | [GitHub](https://github.com/adfinis)
