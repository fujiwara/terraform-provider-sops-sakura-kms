---
page_title: "sops_external Data Source - sops-sakura-kms"
description: |-
  Decrypt SOPS-encrypted content from an external source using Sakura Cloud KMS.
---

# sops_external (Data Source)

Decrypt SOPS-encrypted content provided as a string. This data source is provided by the embedded [carlpett/sops](https://registry.terraform.io/providers/carlpett/sops/latest/docs) provider.

## Example Usage

```hcl
data "sops_external" "secrets" {
  source     = file("secrets.enc.json")
  input_type = "json"
}

output "secret_value" {
  value     = data.sops_external.secrets.data["key"]
  sensitive = true
}
```

## Schema

### Required

- `source` (String) - SOPS-encrypted content as a string.

### Optional

- `input_type` (String) - Type of the input content (`yaml`, `json`, `raw`). Defaults to `json`.

### Read-Only

- `data` (Map of String) - Map of decrypted key-value pairs.
- `raw` (String) - Raw decrypted content.
