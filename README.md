# terraform-provider-sops-sakura-kms

A Terraform provider for decrypting [SOPS](https://github.com/getsops/sops)-encrypted files using [Sakura Cloud KMS](https://manual.sakura.ad.jp/cloud/appliance/kms/index.html).

This provider embeds the [carlpett/sops](https://github.com/carlpett/terraform-provider-sops) data sources and starts a [sops-sakura-kms](https://github.com/fujiwara/sops-sakura-kms) Vault Transit compatible server in-process. No separate provider or background process is needed.

## How it works

1. The provider starts a local Vault Transit Engine compatible HTTP server during `Configure`
2. SOPS decryption runs in the same provider process, connecting to the in-process server
3. Data sources (`sops_file`, `sops_external`) and ephemeral resources (`sops_file`) are provided directly

## Prerequisites

Set the following environment variables:

```bash
export SAKURACLOUD_ACCESS_TOKEN="your-access-token"
export SAKURACLOUD_ACCESS_TOKEN_SECRET="your-access-token-secret"
```

## Usage

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

### With Ephemeral Resources (Terraform v1.11+)

Ephemeral resources prevent secret values from being stored in Terraform state.

```hcl
ephemeral "sops_file" "secrets" {
  source_file = "secrets.enc.yaml"
}
```

## Provider Configuration

| Attribute     | Type   | Required | Default           | Description                                |
|---------------|--------|----------|-------------------|--------------------------------------------|
| `key_id`      | string | Yes      |                   | Sakura Cloud KMS resource ID (12-digit)    |
| `server_addr` | string | No       | `127.0.0.1:8200`  | Address for the local Vault-compatible server |

## Data Sources

This provider includes the same data sources as [carlpett/sops](https://registry.terraform.io/providers/carlpett/sops/latest/docs):

- `sops_file` - Decrypt a SOPS-encrypted file
- `sops_external` - Decrypt SOPS-encrypted content from an external source

See the [carlpett/sops documentation](https://registry.terraform.io/providers/carlpett/sops/latest/docs) for data source attributes.

## Install from source

```bash
make install
```

To use the locally installed provider, add `dev_overrides` to `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "fujiwara/sops-sakura-kms" = "/path/to/your/go/bin"
  }
  direct {}
}
```

## License

MIT
