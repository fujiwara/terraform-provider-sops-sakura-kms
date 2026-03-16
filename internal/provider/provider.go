package provider

import (
	"context"
	"os"

	sopsProvider "github.com/carlpett/terraform-provider-sops/sops"
	ssk "github.com/fujiwara/sops-sakura-kms"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &sskProvider{}
var _ provider.ProviderWithEphemeralResources = &sskProvider{}

// runServerFunc is a function type that starts the server and returns env vars, shutdown func, and error.
type runServerFunc func(ctx context.Context, addr, keyID string) (map[string]string, func(context.Context) error, error)

type sskProvider struct {
	envVars       map[string]string
	shutdown      func(context.Context) error
	runServerFunc runServerFunc
}

type sskProviderModel struct {
	KeyID      types.String `tfsdk:"key_id"`
	ServerAddr types.String `tfsdk:"server_addr"`
}

// New creates a new provider that uses Sakura Cloud KMS.
func New() provider.Provider {
	return &sskProvider{
		runServerFunc: ssk.RunServer,
	}
}

// NewWithCipher creates a new provider that uses the given Cipher for testing.
func NewWithCipher(cipher ssk.Cipher) provider.Provider {
	return &sskProvider{
		runServerFunc: func(ctx context.Context, addr, keyID string) (map[string]string, func(context.Context) error, error) {
			return ssk.RunServerWithCipher(ctx, addr, keyID, cipher)
		},
	}
}

func (p *sskProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "sops"
}

func (p *sskProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides SOPS decryption using Sakura Cloud KMS. " +
			"Starts a Vault Transit Engine compatible server and provides data sources for decrypting SOPS-encrypted files.",
		Attributes: map[string]schema.Attribute{
			"key_id": schema.StringAttribute{
				Description: "Sakura Cloud KMS resource ID (12-digit number).",
				Required:    true,
			},
			"server_addr": schema.StringAttribute{
				Description: "Address for the local Vault-compatible server. Defaults to 127.0.0.1:8200.",
				Optional:    true,
			},
		},
	}
}

func (p *sskProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config sskProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keyID := config.KeyID.ValueString()
	serverAddr := "127.0.0.1:8200"
	if !config.ServerAddr.IsNull() && !config.ServerAddr.IsUnknown() {
		serverAddr = config.ServerAddr.ValueString()
	}

	addEnv, shutdown, err := p.runServerFunc(ctx, serverAddr, keyID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to start sops-sakura-kms server", err.Error())
		return
	}
	p.shutdown = shutdown
	p.envVars = addEnv

	// Set environment variables so that go-sops (running in this process)
	// can connect to the Vault Transit compatible server.
	for k, v := range addEnv {
		os.Setenv(k, v)
	}
}

func (p *sskProvider) Resources(_ context.Context) []func() resource.Resource {
	return nil
}

func (p *sskProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	// Include data sources from carlpett/sops (sops_file, sops_external)
	sp := &sopsProvider.SopsProvider{}
	return sp.DataSources(ctx)
}

func (p *sskProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	// Include ephemeral resources from carlpett/sops (sops_file)
	sp := &sopsProvider.SopsProvider{}
	return sp.EphemeralResources(ctx)
}
