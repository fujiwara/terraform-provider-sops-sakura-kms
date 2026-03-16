package provider_test

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/fujiwara/terraform-provider-sops-sakura-kms/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// mockCipher is a mock implementation of Cipher interface for testing.
type mockCipher struct{}

func (m *mockCipher) Encrypt(_ context.Context, _ string, plaintext []byte) (string, error) {
	return base64.StdEncoding.EncodeToString(plaintext), nil
}

func (m *mockCipher) Decrypt(_ context.Context, _ string, ciphertext string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(ciphertext)
}

func protoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"sopssakurakms": providerserver.NewProtocol6WithError(provider.NewWithCipher(&mockCipher{})),
	}
}

func TestProviderSchema(t *testing.T) {
	server, err := providerserver.NewProtocol6WithError(provider.New())()
	if err != nil {
		t.Fatalf("failed to create provider server: %s", err)
	}
	if server == nil {
		t.Fatal("provider server is nil")
	}
}
