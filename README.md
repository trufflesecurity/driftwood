# Driftwood

Driftwood is a CLI tool that can enable you to lookup whether a private key is used for TLS or as a GitHub SSH key by a user. It does so by computing the public key, so the private key never leaves where you run the tool. Additionally it supports some basic password cracking for encrypted keys.

## Installation

Download the binary from the releases page or build yourself:

```bash
go install github.com/trufflesecurity/driftwood@latest
```

## Usage

Minimal usage is

```bash
$ driftwood path/to/privatekey.pem
```

Run with `--help` to see more options.

## Library Usage

Packages under `pkg/` are libraries that can be used for external consumption. Packages under `pkg/exp/` are considered to be experimental status and may have breaking changes.
