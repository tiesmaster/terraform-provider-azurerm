package keyvault

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

//func NewEphemeralSecrets(_ context.Context) (ephemeral.EphemeralResource, error) {
//	return &ephemeralSecrets{}, nil
//}

//type ephemeralSecrets ephemeral.EphemeralResource

type KeyVaultSecretEphemeralResource struct {
	ephemeral.EphemeralResource
}

type KeyVaultSecretEphemeralResourceModel struct {
	Name       types.String `tfsdk:"name"`
	KeyVaultID types.String `tfsdk:"key_vault_id"`
	Value      types.String `tfsdk:"value"`
	Version    types.String `tfsdk:"version"`
}

func (e KeyVaultSecretEphemeralResource) Metadata(_ context.Context, _ ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = "azurerm_key_vault_secret"
}

func (e KeyVaultSecretEphemeralResource) Schema(ctx context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:   true,
				Validators: nil, // TODO
			},

			"key_vault_id": schema.StringAttribute{
				Required:   true,
				Validators: nil, // TODO
			},

			"value": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},

			"version": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (e KeyVaultSecretEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data KeyVaultSecretEphemeralResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

}
