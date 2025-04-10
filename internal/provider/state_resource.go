// Copyright (c) Christopher Barnes <christopher@barnes.biz>
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	roger "roger/internal/client"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &stateResource{}
	_ resource.ResourceWithConfigure   = &stateResource{}
	_ resource.ResourceWithImportState = &stateResource{}
)

func NewStateResource() resource.Resource {
	return &stateResource{}
}

type stateResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Hostname    types.String `tfsdk:"hostname"`
	Message     types.String `tfsdk:"message"`
	AppState    types.String `tfsdk:"appstate"`
	LastUpdated types.String `tfsdk:"last_updated"`
}

type stateResource struct {
	client *roger.Client
}

func (r *stateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_state"
}

func (r *stateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an roger state.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Numeric identifier of the state.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update of the state.",
				Computed:    true,
			},
			"hostname": schema.StringAttribute{
				Description: "Name of the hostname that belongs to the state.",
				Required:    true,
			},
			"message": schema.StringAttribute{
				Description: "Alert Message",
				Optional:    true,
			},
			"appstate": schema.StringAttribute{
				Description: "Set to 'production', 'draining' or 'quiesce' which are current valid states. Has no effect on alarm status, but may be used to set application state.",
				Required:    true,
			},
		},
	}
}

func (r *stateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan stateResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state, err := r.client.CreateState(plan.Hostname.ValueString(), plan.Message.ValueString(), plan.AppState.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating state",
			"Could not create state, unexpected error: "+err.Error(),
		)
		return
	}

	plan.AppState = types.StringValue(state.AppState)
	plan.Hostname = types.StringValue(state.Hostname)
	plan.ID = types.StringValue(plan.Hostname.ValueString())
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	plan.Message = types.StringValue(state.Message)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *stateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var readState stateResourceModel
	diags := req.State.Get(ctx, &readState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state, err := r.client.GetState(readState.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading roger state",
			"Could not read roger state ID "+readState.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	readState.Hostname = types.StringValue(state.Hostname)

	diags = resp.State.Set(ctx, &readState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *stateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan stateResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.UpdateState(plan.Hostname.ValueString(), plan.Message.ValueString(), plan.AppState.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating roger state",
			"Could not update state, unexpected error: "+err.Error(),
		)
		return
	}

	statePtr, err := r.client.GetState(plan.Hostname.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading roger state",
			"Could not read roger state Hostname "+plan.Hostname.ValueString()+": "+err.Error(),
		)
		return
	}

	plan.Hostname = types.StringValue(statePtr.Hostname)
	plan.Message = types.StringValue(statePtr.Message)
	plan.AppState = types.StringValue(statePtr.AppState)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *stateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state stateResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteState(state.Hostname.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting roger state",
			"Could not delete state, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *stateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*roger.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *roger.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *stateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
