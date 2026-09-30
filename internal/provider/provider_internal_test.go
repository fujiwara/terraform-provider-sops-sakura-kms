package provider

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

type testCipher struct{}

func (testCipher) Encrypt(_ context.Context, _ string, plaintext []byte) (string, error) {
	return base64.StdEncoding.EncodeToString(plaintext), nil
}

func (testCipher) Decrypt(_ context.Context, _ string, ciphertext string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(ciphertext)
}

func TestResolveServerAddr(t *testing.T) {
	tests := []struct {
		name string
		in   types.String
		want string
	}{
		{"null", types.StringNull(), defaultServerAddr},
		{"unknown", types.StringUnknown(), defaultServerAddr},
		{"empty", types.StringValue(""), defaultServerAddr},
		{"explicit", types.StringValue("127.0.0.1:18200"), "127.0.0.1:18200"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveServerAddr(tt.in); got != tt.want {
				t.Errorf("resolveServerAddr() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestRunServerDefaultAddrConcurrent verifies that multiple servers can run at
// once with the default address, and SOPS files are encrypted with 127.0.0.1:8200.
func TestRunServerDefaultAddrConcurrent(t *testing.T) {
	ctx := context.Background()
	p := NewWithCipher(testCipher{}).(*sskProvider)
	agentAddrs := map[string]bool{}
	for range 2 {
		env, shutdown, err := p.runServerFunc(ctx, resolveServerAddr(types.StringNull()), "123456789012")
		if err != nil {
			t.Fatalf("failed to run server: %s", err)
		}
		t.Cleanup(func() { shutdown(context.Background()) })

		agentAddr := env["VAULT_AGENT_ADDR"]
		if !strings.HasPrefix(agentAddr, "http://127.0.0.1:") || agentAddr == "http://127.0.0.1:8200" {
			t.Errorf("unexpected VAULT_AGENT_ADDR: %q", agentAddr)
		}
		agentAddrs[agentAddr] = true
		if got, want := env["SOPS_VAULT_URIS"], "http://127.0.0.1:8200/v1/transit/encrypt/123456789012"; got != want {
			t.Errorf("SOPS_VAULT_URIS = %q, want %q", got, want)
		}

		resp, err := http.Get(agentAddr + "/health")
		if err != nil {
			t.Fatalf("failed to request health check: %s", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("health check status = %d", resp.StatusCode)
		}
	}
	if len(agentAddrs) != 2 {
		t.Errorf("servers should listen on different addresses: %v", agentAddrs)
	}
}
