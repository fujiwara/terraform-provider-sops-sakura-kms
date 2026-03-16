package main

import (
	"context"
	"log"

	"github.com/fujiwara/terraform-provider-sops-sakura-kms/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

func main() {
	err := providerserver.Serve(context.Background(), provider.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/fujiwara/sops-sakura-kms",
	})
	if err != nil {
		log.Fatal(err)
	}
}
