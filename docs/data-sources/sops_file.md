---
page_title: "sops_file Data Source - sops-sakura-kms"
description: |-
  Decrypt a SOPS-encrypted file using Sakura Cloud KMS.
---

# sops_file (Data Source)

Decrypt a SOPS-encrypted file. This data source is provided by the embedded [carlpett/sops](https://registry.terraform.io/providers/carlpett/sops/latest/docs) provider.

## Example Usage

```hcl
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

- `source_file` (String) - Path to the SOPS-encrypted file.

### Optional

- `input_type` (String) - Type of the input file (`yaml`, `json`, `raw`). If not specified, the file extension is used.

### Read-Only

- `data` (Map of String) - Map of decrypted key-value pairs from the file.
- `raw` (String) - Raw decrypted content of the file.
