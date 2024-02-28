// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ApplicationDataSource{}

func NewApplicationDataSource() datasource.DataSource {
	return &ApplicationDataSource{}
}

// ApplicationDataSource defines the data source implementation.
type ApplicationDataSource struct {
	client *http.Client
}

// ApplicationDataSourceModel describes the data source data model.
type ApplicationDataSourceModel struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Priority    types.String `tfsdk:"priority"`
	Id          types.String `tfsdk:"id"`
	Token       types.String `tfsdk:"token"`
}

func (d *ApplicationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (d *ApplicationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Application data source",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the gotify application you want to create",
				Optional:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the gotify application",
				Optional:            true,
			},
			"priority": schema.StringAttribute{
				MarkdownDescription: "Priority of the application",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Application identifier",
			},
			"token": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Application identifier",
			},
		},
	}
}

func (d *ApplicationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*http.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *ApplicationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ApplicationDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	url := strings.Trim(Config.Url.String(), "\"")
	token := strings.Trim(Config.Token.String(), "\"")
	id := strings.Trim(data.Id.String(), "\"")

	httpReq, err := http.NewRequest("GET", url+"/application", nil)
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Can't send request to Gotify", err.Error())
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Gotify-Key", token)

	httpRes, err := d.client.Do(httpReq)
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("API Error when contacting Gotify instance", err.Error())
		return
	}
	defer httpRes.Body.Close()

	statusCode := httpRes.StatusCode

	if statusCode == 401 {
		bodyBytes, _ := io.ReadAll(httpRes.Body)
		bodyString := string(bodyBytes)

		resp.Diagnostics.AddError("Not Allowed", fmt.Sprintf("Bad token (?) : %s", bodyString))
		return
	} else if statusCode != 200 {
		bodyBytes, _ := io.ReadAll(httpRes.Body)
		bodyString := string(bodyBytes)

		resp.Diagnostics.AddError("API Error when contacting Gotify instance", fmt.Sprintf("Received a %s response code : %s", strconv.Itoa(statusCode), bodyString))
		return
	}

	type JsonReponse []struct {
		DefaultPriority int64       `json:"defaultPriority"`
		Description     string      `json:"description"`
		ID              int64       `json:"id"`
		Image           string      `json:"image"`
		Internal        bool        `json:"internal"`
		LastUsed        interface{} `json:"lastUsed"`
		Name            string      `json:"name"`
		Token           string      `json:"token"`
	}

	var respData JsonReponse

	err = json.NewDecoder(httpRes.Body).Decode(&respData)
	if err != nil {
		resp.Diagnostics.AddError("API Error when contacting Gotify instance", err.Error())
		return
	}

	tflog.Info(ctx, fmt.Sprintf("Searched id: %s", id))

	ok := false
	for _, Application := range respData {
		if strconv.Itoa(int(Application.ID)) == id {
			ok = true
			data.Name = types.StringValue(Application.Name)
			data.Description = types.StringValue(Application.Description)
			data.Id = types.StringValue(strconv.FormatInt(Application.ID, 10))
			data.Priority = types.StringValue(strconv.FormatInt(Application.DefaultPriority, 10))
			data.Token = types.StringValue(Application.Token)
		}
	}

	if !ok {
		resp.Diagnostics.AddError("API Error", "No application found with this id")
		return
	}

	// data.Description = types.StringValue("Description")
	// data.Id = types.StringValue("Application-id")
	// data.Priority = types.StringValue("Priority")
	// data.Token = types.StringValue("Token")
	// data.Name = types.StringValue("Name")

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
