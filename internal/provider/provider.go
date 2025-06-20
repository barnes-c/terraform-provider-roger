// SPDX-FileCopyrightText: 2025 CERN
//
// SPDX-License-Identifier: GPL-3.0-or-later

package provider

import (
	"context"
	"os"
	roger "roger/internal/client"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ provider.Provider = &rogerProvider{}
)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &rogerProvider{
			version: version,
		}
	}
}

type rogerProviderModel struct {
	Host types.String `tfsdk:"host"`
	Port types.Number `tfsdk:"port"`
}

type rogerProvider struct {
	version string
}

func (p *rogerProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "roger"
	resp.Version = p.version
}

func (p *rogerProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with roger.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "URI for roger API. May also be provided via ROGER_HOST environment variable.",
				Optional:    true,
			},
			"port": schema.NumberAttribute{
				Description: "Port for roger API. May also be provided via ROGER_PORT environment variable.",
				Optional:    true,
			},
		},
	}
}

func (p *rogerProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring roger client")

	var config rogerProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown roger API Host",
			"The provider cannot create the roger API client as there is an unknown configuration value for the roger host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ROGER_HOST environment variable.",
		)
	}

	if config.Port.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("port"),
			"Unknown roger host Port",
			"The provider cannot create the roger API client as there is an unknown configuration value for the roger port. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ROGER_PORT environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	host := os.Getenv("ROGER_HOST")
	if host == "" {
		host = "woger-direct.cern.ch"
	}

	port := 8201
	if portStr := os.Getenv("ROGER_PORT"); portStr != "" {
		if parsed, err := strconv.Atoi(portStr); err == nil {
			port = parsed
		}

	}

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.Port.IsNull() {
		bf := config.Port.ValueBigFloat()
		portInt64, _ := bf.Int64()
		port = int(portInt64)
	}

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing roger API Host",
			"The provider cannot create the roger API client as there is a missing or empty value for the roger host. "+
				"Set the host value in the configuration or use the ROGER_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if port == 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("port"),
			"Missing roger Port",
			"The provider cannot create the roger API client as there is a missing or empty value for the roger port. "+
				"Set the port value in the configuration or use the ROGER_PORT environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "roger_host", host)
	ctx = tflog.SetField(ctx, "roger_port", port)

	tflog.Debug(ctx, "Creating roger client")

	client, err := roger.NewClient(host, port)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create roger API Client",
			"An unexpected error occurred when creating the roger API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"roger Client Error: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured roger client", map[string]any{"success": true})
}

func (p *rogerProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewStateResource,
	}
}

func (p *rogerProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
