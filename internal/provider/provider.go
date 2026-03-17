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
	"github.com/sacloud/saclient-go"
)

var _ provider.Provider = &sskProvider{}
var _ provider.ProviderWithEphemeralResources = &sskProvider{}

// runServerFunc is a function type that starts the server and returns env vars, shutdown func, and error.
type runServerFunc func(ctx context.Context, addr, keyID string, opts ...ssk.Option) (map[string]string, func(context.Context) error, error)

type sskProvider struct {
	envVars       map[string]string
	shutdown      func(context.Context) error
	runServerFunc runServerFunc
}

type sskProviderModel struct {
	KeyID                 types.String `tfsdk:"key_id"`
	ServerAddr            types.String `tfsdk:"server_addr"`
	Profile               types.String `tfsdk:"profile"`
	Token                 types.String `tfsdk:"token"`
	Secret                types.String `tfsdk:"secret"`
	ServicePrincipalID    types.String `tfsdk:"service_principal_id"`
	ServicePrincipalKeyID types.String `tfsdk:"service_principal_key_id"`
	PrivateKeyPath        types.String `tfsdk:"private_key_path"`
	APIRootURL            types.String `tfsdk:"api_root_url"`
	RetryMax              types.Int64  `tfsdk:"retry_max"`
	APIRequestTimeout     types.Int64  `tfsdk:"api_request_timeout"`
	APIRequestRateLimit   types.Int64  `tfsdk:"api_request_rate_limit"`
	Trace                 types.String `tfsdk:"trace"`
}

// Verify that sskProviderModel implements saclient.TerraformProviderInterface.
var _ saclient.TerraformProviderInterface = &sskProviderModel{}

func (m *sskProviderModel) LookupClientConfigProfileName() (string, bool) {
	return lookupString(m.Profile)
}

func (m *sskProviderModel) LookupClientConfigServicePrincipalID() (string, bool) {
	return lookupString(m.ServicePrincipalID)
}

func (m *sskProviderModel) LookupClientConfigServicePrincipalKeyID() (string, bool) {
	return lookupString(m.ServicePrincipalKeyID)
}

func (m *sskProviderModel) LookupClientConfigPrivateKeyPath() (string, bool) {
	return lookupString(m.PrivateKeyPath)
}

func (m *sskProviderModel) LookupClientConfigAccessToken() (string, bool) {
	return lookupString(m.Token)
}

func (m *sskProviderModel) LookupClientConfigAccessTokenSecret() (string, bool) {
	return lookupString(m.Secret)
}

func (m *sskProviderModel) LookupClientConfigZone() (string, bool) {
	return "", false
}

func (m *sskProviderModel) LookupClientConfigDefaultZone() (string, bool) {
	return "", false
}

func (m *sskProviderModel) LookupClientConfigZones() ([]string, bool) {
	return nil, false
}

func (m *sskProviderModel) LookupClientConfigRetryMax() (int64, bool) {
	return lookupInt64(m.RetryMax)
}

func (m *sskProviderModel) LookupClientConfigRetryWaitMax() (int64, bool) {
	return 0, false
}

func (m *sskProviderModel) LookupClientConfigRetryWaitMin() (int64, bool) {
	return 0, false
}

func (m *sskProviderModel) LookupClientConfigAPIRootURL() (string, bool) {
	return lookupString(m.APIRootURL)
}

func (m *sskProviderModel) LookupClientConfigAPIRequestTimeout() (int64, bool) {
	return lookupInt64(m.APIRequestTimeout)
}

func (m *sskProviderModel) LookupClientConfigAPIRequestRateLimit() (int64, bool) {
	return lookupInt64(m.APIRequestRateLimit)
}

func (m *sskProviderModel) LookupClientConfigTraceMode() (string, bool) {
	return lookupString(m.Trace)
}

func lookupString(v types.String) (string, bool) {
	if v.IsNull() || v.IsUnknown() {
		return "", false
	}
	return v.ValueString(), true
}

func lookupInt64(v types.Int64) (int64, bool) {
	if v.IsNull() || v.IsUnknown() {
		return 0, false
	}
	return v.ValueInt64(), true
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
		runServerFunc: func(ctx context.Context, addr, keyID string, opts ...ssk.Option) (map[string]string, func(context.Context) error, error) {
			return ssk.RunServer(ctx, addr, keyID, ssk.WithCipher(cipher))
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
			"profile": schema.StringAttribute{
				Description: "Profile name for shared credentials (~/.usacloud/<profile>/config.json).",
				Optional:    true,
			},
			"token": schema.StringAttribute{
				Description: "API access token. Can also be set via SAKURA_ACCESS_TOKEN environment variable.",
				Optional:    true,
			},
			"secret": schema.StringAttribute{
				Description: "API access token secret. Can also be set via SAKURA_ACCESS_TOKEN_SECRET environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"service_principal_id": schema.StringAttribute{
				Description: "Service principal ID for authentication.",
				Optional:    true,
			},
			"service_principal_key_id": schema.StringAttribute{
				Description: "Service principal key ID for authentication.",
				Optional:    true,
			},
			"private_key_path": schema.StringAttribute{
				Description: "Path to the private key file for service principal authentication.",
				Optional:    true,
			},
			"api_root_url": schema.StringAttribute{
				Description: "Custom API root URL.",
				Optional:    true,
			},
			"retry_max": schema.Int64Attribute{
				Description: "Maximum number of API call retries.",
				Optional:    true,
			},
			"api_request_timeout": schema.Int64Attribute{
				Description: "API request timeout in seconds.",
				Optional:    true,
			},
			"api_request_rate_limit": schema.Int64Attribute{
				Description: "Maximum API calls per second.",
				Optional:    true,
			},
			"trace": schema.StringAttribute{
				Description: "Enable API trace logging.",
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

	// Build saclient.Client from provider config
	var sc saclient.Client
	sc.SetEnviron(os.Environ())
	if err := sc.SettingsFromTerraformProvider(&config); err != nil {
		resp.Diagnostics.AddError("Failed to configure saclient", err.Error())
		return
	}
	if err := sc.Populate(); err != nil {
		resp.Diagnostics.AddError("Failed to populate saclient", err.Error())
		return
	}

	addEnv, shutdown, err := p.runServerFunc(ctx, serverAddr, keyID, ssk.WithClient(&sc))
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
