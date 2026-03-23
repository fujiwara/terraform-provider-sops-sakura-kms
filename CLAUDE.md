# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Terraform provider that decrypts SOPS-encrypted files using Sakura Cloud KMS. It embeds `carlpett/terraform-provider-sops` and runs an in-process Vault Transit Engine-compatible server via `fujiwara/sops-sakura-kms` for KMS operations.

## Build & Test Commands

```bash
make build          # Build binary
make test           # Run tests (go test -v ./...)
make install        # Install to $GOPATH/bin
make clean          # Remove binary

# Run a single test
go test -v -run TestProviderSchema ./internal/provider/

# CI runs with race detection
go test -race ./...
```

## Architecture

- **`main.go`** — Entry point. Starts the provider server registered at `registry.terraform.io/fujiwara/sops-sakura-kms`.
- **`internal/provider/provider.go`** — Core provider implementation (`sskProvider`).
  - `New()` creates the production provider; `NewWithCipher()` injects a mock cipher for testing.
  - `Configure()` starts an in-process Vault Transit-compatible HTTP server and sets environment variables for SOPS to connect to it.
  - Data sources (`sops_file`, `sops_external`) and ephemeral resources (`sops_file`) are delegated to the embedded `carlpett/terraform-provider-sops`.
  - No custom Terraform resources are defined.
- **`internal/provider/provider_test.go`** — Tests use a `mockCipher` (base64 encode/decode) to avoid requiring real Sakura Cloud credentials.

## Key Design Decisions

- **Adapter pattern**: Wraps `carlpett/terraform-provider-sops` with Sakura Cloud KMS support rather than reimplementing SOPS data sources.
- **Dependency injection**: `runServerFunc` field enables test isolation without real KMS calls.
- **In-process server**: The Vault Transit-compatible server runs in the same process — no external service needed.

## Registry Documentation

Documentation for the Terraform Registry is in `docs/`. Since `tfplugindocs` cannot auto-generate docs for this provider (data sources are delegated from `carlpett/sops` with a different type name prefix), docs must be maintained manually. When adding or changing data sources, ephemeral resources, or provider schema, update the corresponding files in `docs/` as well.

## Provider Configuration

Requires environment variables `SAKURA_ACCESS_TOKEN` and `SAKURA_ACCESS_TOKEN_SECRET` for real KMS operations.

Provider attributes:
- `key_id` (optional): 12-digit Sakura Cloud KMS resource ID. Not required for decryption (key ID is read from SOPS file metadata)
- `server_addr` (optional, default `127.0.0.1:8200`): Address for local Vault-compatible server
