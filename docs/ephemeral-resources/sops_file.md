---
page_title: "sops_file Ephemeral Resource - sops-sakura-kms"
description: |-
  Decrypt a SOPS-encrypted file without storing secrets in state.
---

# sops_file (Ephemeral Resource)

Decrypt a SOPS-encrypted file without storing secret values in Terraform state. Requires Terraform v1.11 or later. This ephemeral resource is provided by the embedded [carlpett/sops](https://registry.terraform.io/providers/carlpett/sops/latest/docs) provider.

## Example Usage

```hcl
ephemeral "sops_file" "secrets" {
  source_file = "secrets.enc.yaml"
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
