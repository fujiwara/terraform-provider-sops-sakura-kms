---
page_title: "Provider: sops-sakura-kms"
description: |-
  Decrypt SOPS-encrypted files using Sakura Cloud KMS.
---

# sops-sakura-kms Provider

A Terraform provider for decrypting [SOPS](https://github.com/getsops/sops)-encrypted files using [Sakura Cloud KMS](https://manual.sakura.ad.jp/cloud/appliance/kms/index.html).

This provider embeds the [carlpett/sops](https://registry.terraform.io/providers/carlpett/sops/latest/docs) data sources and starts a [sops-sakura-kms](https://github.com/fujiwara/sops-sakura-kms) Vault Transit compatible server in-process. No separate provider or background process is needed.

SOPS files encrypted with non-Sakura Cloud KMS keys (e.g., AWS KMS, GCP KMS, Azure Key Vault, PGP) can also be decrypted without any Sakura Cloud credentials. SOPS resolves the appropriate key provider from the file metadata.

## Authentication

Sakura Cloud credentials are only required when decrypting files encrypted with Sakura Cloud KMS. The provider supports the same authentication methods as [terraform-provider-sakura](https://registry.terraform.io/providers/sacloud/sakura/latest/docs). The priority order is: HCL attributes > environment variables > profile.

### Environment variables

```bash
export SAKURA_ACCESS_TOKEN="your-access-token"
export SAKURA_ACCESS_TOKEN_SECRET="your-access-token-secret"
```

### Provider attributes

```hcl
provider "sops" {
  token  = "your-access-token"
  secret = "your-access-token-secret"
}
```

### Profile

```hcl
provider "sops" {
  profile = "your-profile"
}
```

Profile reads credentials from `~/.usacloud/<profile>/config.json`.

## Example Usage

```hcl
terraform {
  required_providers {
    sops = {
      source = "fujiwara/sops-sakura-kms"
    }
  }
}

provider "sops" {
}

data "sops_file" "secrets" {
  source_file = "secrets.enc.yaml"
}

output "secret_value" {
  value     = data.sops_file.secrets.data["key"]
  sensitive = true
}
```

## Schema

### Optional

- `key_id` (String) - Sakura Cloud KMS resource ID (12-digit number). Optional: not required for decryption as key ID is read from SOPS file metadata. When set, `SOPS_VAULT_URIS` is configured automatically.
- `server_addr` (String) - Address for the local Vault-compatible server. Defaults to `127.0.0.1:8200`.
- `profile` (String) - Profile name for shared credentials (`~/.usacloud/<profile>/config.json`).
- `token` (String) - API access token. Can also be set via `SAKURA_ACCESS_TOKEN` environment variable.
- `secret` (String, Sensitive) - API access token secret. Can also be set via `SAKURA_ACCESS_TOKEN_SECRET` environment variable.
- `service_principal_id` (String) - Service principal ID for authentication.
- `service_principal_key_id` (String) - Service principal key ID for authentication.
- `private_key_path` (String) - Path to the private key file for service principal authentication.
- `api_root_url` (String) - Custom API root URL.
- `retry_max` (Number) - Maximum number of API call retries.
- `api_request_timeout` (Number) - API request timeout in seconds.
- `api_request_rate_limit` (Number) - Maximum API calls per second.
- `trace` (String) - Enable API trace logging.
