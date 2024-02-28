// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure GotifyProvider satisfies various provider interfaces.
var _ provider.Provider = &GotifyProvider{}

// GotifyProvider defines the provider implementation.
type GotifyProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// GotifyProviderModel describes the provider data model.
type GotifyProviderModel struct {
	Token types.String `tfsdk:"token"`
	Url   types.String `tfsdk:"url"`
}

// variable contains provider configuration
var Config GotifyProviderModel

func (p *GotifyProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "gotify"
	resp.Version = p.version
}

func (p *GotifyProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				MarkdownDescription: "Token of Gotify Client",
				Required:            true,
				Optional:            false,
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "URL for Gotify Instance",
				Required:            true,
			},
		},
	}
}

func (p *GotifyProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data GotifyProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	url := strings.Trim(data.Url.String(), "\"")
	token := strings.Trim(data.Token.String(), "\"")
	// priority := data.Priority
	client := http.DefaultClient

	httpReq, err := http.NewRequest("GET", url+"/application", nil)
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("API Error when contacting Gotify instance", err.Error())
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Gotify-Key", token)

	httpRes, err := client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Can't contact Gotify Instance", err.Error())
		return
	}

	defer httpRes.Body.Close()

	statusCode := httpRes.StatusCode

	if statusCode == 401 {
		resp.Diagnostics.AddError("Not Allowed", "Bad token (?)")
		return
	} else if statusCode != 200 {
		resp.Diagnostics.AddError("API Error when contacting Gotify instance", "Received a non-200 response code")
		return
	}

	Config = data

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *GotifyProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewApplicationResource,
	}
}

func (p *GotifyProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewApplicationDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &GotifyProvider{
			version: version,
		}
	}
}
