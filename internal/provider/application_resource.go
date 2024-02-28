// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ApplicationResource{}
var _ resource.ResourceWithImportState = &ApplicationResource{}

func NewApplicationResource() resource.Resource {
	return &ApplicationResource{}
}

// ApplicationResource defines the resource implementation.
type ApplicationResource struct {
	client *http.Client
}

// ApplicationResourceModel describes the resource data model.
type ApplicationResourceModel struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Priority    types.String `tfsdk:"priority"`
	Id          types.String `tfsdk:"id"`
	Token       types.String `tfsdk:"token"`
}

func (r *ApplicationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (r *ApplicationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Application resource for gotify",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the gotify application you want to create",
				Optional:            false,
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the gotify application",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("Description not configured"),
			},
			"priority": schema.StringAttribute{
				MarkdownDescription: "Priority of the application",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("1"),
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Application identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"token": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Application identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ApplicationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.

	if req.ProviderData == nil {
		tflog.Info(ctx, "No informations provided")
		return
	}

	client, ok := req.ProviderData.(*http.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *ApplicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ApplicationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	url := strings.Trim(Config.Url.String(), "\"")
	token := strings.Trim(Config.Token.String(), "\"")

	priority, err := strconv.Atoi(strings.Trim(data.Priority.String(), "\""))
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Priority cannot be parsed as Int", err.Error())
		return
	}

	reqData := map[string]interface{}{
		"defaultPriority": priority,
		"description":     strings.Trim(data.Description.String(), "\""),
		"name":            strings.Trim(data.Name.String(), "\""),
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Can't convert data to json", err.Error())
		return
	}

	httpReq, err := http.NewRequest("POST", url+"/application", bytes.NewBuffer(jsonData))
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Can't send request to Gotify", err.Error())
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Gotify-Key", token)

	httpRes, err := r.client.Do(httpReq)
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("API Error when contacting Gotify instance", err.Error())
		return
	}
	defer httpRes.Body.Close()

	statusCode := httpRes.StatusCode

	if statusCode == 401 {
		bodyBytes, _ := ioutil.ReadAll(httpRes.Body)
		bodyString := string(bodyBytes)

		resp.Diagnostics.AddError("Not Allowed", fmt.Sprintf("Bad token (?) : %s", bodyString))
		return
	} else if statusCode != 200 {
		bodyBytes, _ := ioutil.ReadAll(httpRes.Body)
		bodyString := string(bodyBytes)

		resp.Diagnostics.AddError("API Error when contacting Gotify instance", fmt.Sprintf("Received a %s response code : %s", strconv.Itoa(statusCode), bodyString))
		return
	}

	type Response struct {
		ID       int    `json:"id"`
		Token    string `json:"token"`
		Internal bool   `json:"internal"`
	}
	var respData Response

	err = json.NewDecoder(httpRes.Body).Decode(&respData)
	if err != nil {
		resp.Diagnostics.AddError("API Error when contacting Gotify instance", "Failed to decode response body")
		return
	}

	data.Id = types.StringValue(strconv.Itoa(respData.ID))
	data.Token = types.StringValue(respData.Token)

	tflog.Info(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ApplicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ApplicationResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *ApplicationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ApplicationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	url := strings.Trim(Config.Url.String(), "\"")
	token := strings.Trim(Config.Token.String(), "\"")
	priority, err := strconv.Atoi(strings.Trim(data.Priority.String(), "\""))
	id := strings.Trim(data.Id.String(), "\"")

	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Priority cannot be parsed as Int", err.Error())
		return
	}

	reqData := map[string]interface{}{
		"defaultPriority": priority,
		"description":     strings.Trim(data.Description.String(), "\""),
		"name":            strings.Trim(data.Name.String(), "\""),
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Can't convert data to json", err.Error())
		return
	}

	httpReq, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s/%s", url, "application", id), bytes.NewBuffer(jsonData))
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Can't send request to Gotify", err.Error())
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Gotify-Key", token)

	httpRes, err := r.client.Do(httpReq)
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("API Error when contacting Gotify instance", err.Error())
		return
	}
	defer httpRes.Body.Close()

	statusCode := httpRes.StatusCode

	if statusCode == 401 {
		bodyBytes, _ := ioutil.ReadAll(httpRes.Body)
		bodyString := string(bodyBytes)

		resp.Diagnostics.AddError("Not Allowed", fmt.Sprintf("Bad token (?) : %s", bodyString))
		return
	} else if statusCode != 200 {
		bodyBytes, _ := ioutil.ReadAll(httpRes.Body)
		bodyString := string(bodyBytes)

		resp.Diagnostics.AddError("API Error when contacting Gotify instance", fmt.Sprintf("Received a %s response code : %s", strconv.Itoa(statusCode), bodyString))
		return
	}

	tflog.Info(ctx, "Updated a resource")

}

func (r *ApplicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ApplicationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := strings.Trim(Config.Url.String(), "\"")
	token := strings.Trim(Config.Token.String(), "\"")
	id := strings.Trim(data.Id.String(), "\"")

	httpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s/%s", url, "application", id), nil)
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Can't send request to Gotify", err.Error())
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Gotify-Key", token)

	httpRes, err := r.client.Do(httpReq)
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("API Error when contacting Gotify instance", err.Error())
		return
	}
	defer httpRes.Body.Close()

	statusCode := httpRes.StatusCode

	if statusCode == 401 {
		bodyBytes, _ := ioutil.ReadAll(httpRes.Body)
		bodyString := string(bodyBytes)

		resp.Diagnostics.AddError("Not Allowed", fmt.Sprintf("Bad token (?) : %s", bodyString))
		return
	} else if statusCode != 200 {
		bodyBytes, _ := ioutil.ReadAll(httpRes.Body)
		bodyString := string(bodyBytes)

		resp.Diagnostics.AddError("API Error when contacting Gotify instance", fmt.Sprintf("Received a %s response code : %s", strconv.Itoa(statusCode), bodyString))
		return
	}

	tflog.Info(ctx, "Deleted a resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}

func (r *ApplicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
