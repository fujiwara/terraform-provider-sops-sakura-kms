---
page_title: "Provider: sops-sakura-kms"
description: |-
  Decrypt SOPS-encrypted files using Sakura Cloud KMS.
---

# sops-sakura-kms Provider

A Terraform provider for decrypting [SOPS](https://github.com/getsops/sops)-encrypted files using [Sakura Cloud KMS](https://manual.sakura.ad.jp/cloud/appliance/kms/index.html).

This provider embeds the [carlpett/sops](https://registry.terraform.io/providers/carlpett/sops/latest/docs) data sources and starts a [sops-sakura-kms](https://github.com/fujiwara/sops-sakura-kms) Vault Transit compatible server in-process. No separate provider or background process is needed.

## Prerequisites

Set the following environment variables:

```bash
export SAKURACLOUD_ACCESS_TOKEN="your-access-token"
export SAKURACLOUD_ACCESS_TOKEN_SECRET="your-access-token-secret"
```

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
  key_id = "123456789012"  # Sakura Cloud KMS resource ID
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

### Required

- `key_id` (String) - Sakura Cloud KMS resource ID (12-digit number).

### Optional

- `server_addr` (String) - Address for the local Vault-compatible server. Defaults to `127.0.0.1:8200`.
