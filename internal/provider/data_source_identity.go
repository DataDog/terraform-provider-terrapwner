// Copyright (c) Datadog, Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &TerrapwnerIdentityDataSource{}

func NewTerrapwnerIdentityDataSource() datasource.DataSource {
	return &TerrapwnerIdentityDataSource{}
}

// TerrapwnerIdentityDataSource defines the data source implementation.
type TerrapwnerIdentityDataSource struct{}

// TerrapwnerIdentityDataSourceModel describes the data source data model.
type TerrapwnerIdentityDataSourceModel struct {
	Id            types.String `tfsdk:"id"`
	CloudProvider types.String `tfsdk:"cloud_provider"` // e.g., "aws", "gcp", "azure"
	AccountId     types.String `tfsdk:"account_id"`     // e.g., AWS account ID
	ResourceId    types.String `tfsdk:"resource_id"`    // e.g., AWS ARN
	CallerName    types.String `tfsdk:"caller_name"`    // e.g., role name or user name
	CallerType    types.String `tfsdk:"caller_type"`    // e.g., "role", "user", "assumed-role"
	SessionName   types.String `tfsdk:"session_name"`   // e.g., session name for assumed roles
	Region        types.String `tfsdk:"region"`         // e.g., AWS region
}

func (d *TerrapwnerIdentityDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_identity"
}

func (d *TerrapwnerIdentityDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves identity information about the entity running Terraform",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Identifier for this data source",
				Computed:            true,
			},
			"cloud_provider": schema.StringAttribute{
				MarkdownDescription: "Cloud provider (e.g., aws, gcp, azure)",
				Computed:            true,
			},
			"account_id": schema.StringAttribute{
				MarkdownDescription: "Cloud account ID (e.g., AWS account ID)",
				Computed:            true,
			},
			"resource_id": schema.StringAttribute{
				MarkdownDescription: "Resource identifier (e.g., AWS ARN)",
				Computed:            true,
			},
			"caller_name": schema.StringAttribute{
				MarkdownDescription: "Name of the caller (e.g., role name or user name)",
				Computed:            true,
			},
			"caller_type": schema.StringAttribute{
				MarkdownDescription: "Type of the caller (e.g., role, user, assumed-role)",
				Computed:            true,
			},
			"session_name": schema.StringAttribute{
				MarkdownDescription: "Session name for assumed roles",
				Computed:            true,
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "Cloud region",
				Computed:            true,
			},
		},
	}
}

func (d *TerrapwnerIdentityDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// No configuration needed
}

func (d *TerrapwnerIdentityDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TerrapwnerIdentityDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Try to detect the cloud provider and environment
	provider, region := d.detectProviderAndEnvironment()

	// Set the provider and region
	data.CloudProvider = types.StringValue(provider)
	data.Region = types.StringValue(region)

	// Get identity information based on the provider
	switch provider {
	case "aws":
		if err := d.getAWSIdentity(ctx, &data); err != nil {
			// Log the error but don't fail the data source
			resp.Diagnostics.AddWarning("Failed to get AWS identity", err.Error())
			// Set default values for AWS-specific fields
			data.AccountId = types.StringValue("unknown")
			data.ResourceId = types.StringValue("unknown")
			data.CallerName = types.StringValue("unknown")
			data.CallerType = types.StringValue("unknown")
			data.SessionName = types.StringValue("unknown")
		}
	case "":
		// No cloud provider detected, set all fields to unknown
		data.AccountId = types.StringValue("unknown")
		data.ResourceId = types.StringValue("unknown")
		data.CallerName = types.StringValue("unknown")
		data.CallerType = types.StringValue("unknown")
		data.SessionName = types.StringValue("unknown")
	default:
		// Unsupported provider detected, set all fields to unknown
		resp.Diagnostics.AddWarning("Unsupported provider", fmt.Sprintf("Provider %s is not supported", provider))
		data.AccountId = types.StringValue("unknown")
		data.ResourceId = types.StringValue("unknown")
		data.CallerName = types.StringValue("unknown")
		data.CallerType = types.StringValue("unknown")
		data.SessionName = types.StringValue("unknown")
	}

	// Set a unique ID for this data source
	data.Id = types.StringValue(fmt.Sprintf("%s-%s", provider, data.AccountId.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *TerrapwnerIdentityDataSource) detectProviderAndEnvironment() (string, string) {
	// Check for AWS environment variables
	if os.Getenv("AWS_ACCESS_KEY_ID") != "" {
		// Get AWS region from environment or default to us-east-1
		region := os.Getenv("AWS_REGION")
		if region == "" {
			region = "us-east-1"
		}
		return "aws", region
	}

	// No cloud provider detected
	return "", ""
}

func (d *TerrapwnerIdentityDataSource) getAWSIdentity(ctx context.Context, data *TerrapwnerIdentityDataSourceModel) error {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(data.Region.ValueString()))
	if err != nil {
		return fmt.Errorf("unable to load AWS configuration: %w", err)
	}

	// Create STS client
	stsClient := sts.NewFromConfig(cfg)

	// Get caller identity
	identity, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return fmt.Errorf("unable to get AWS identity: %w", err)
	}

	// Set basic identity information
	data.AccountId = types.StringValue(*identity.Account)
	data.ResourceId = types.StringValue(*identity.Arn)

	// Parse the ARN using AWS SDK
	parsedARN, err := arn.Parse(*identity.Arn)
	if err != nil {
		return fmt.Errorf("unable to parse ARN: %w", err)
	}

	// Extract resource type and ID from the resource string
	resourceParts := strings.Split(parsedARN.Resource, "/")
	if len(resourceParts) < 2 {
		return fmt.Errorf("invalid resource format in ARN: %s", parsedARN.Resource)
	}
	resourceType := resourceParts[0]
	resourceID := strings.Join(resourceParts[1:], "/")

	// Determine caller type and name from ARN
	callerType, callerName, sessionName := d.determineCallerInfo(parsedARN.Service, resourceType, resourceID)
	data.CallerType = types.StringValue(callerType)
	data.CallerName = types.StringValue(callerName)
	data.SessionName = types.StringValue(sessionName)

	return nil
}

// determineCallerInfo extracts the caller type and name from an AWS ARN
func (d *TerrapwnerIdentityDataSource) determineCallerInfo(service, resourceType, resourceID string) (string, string, string) {
	switch {
	case service == "sts" && resourceType == "assumed-role":
		// For assumed roles, resourceID is in format "role-name/session-name"
		parts := strings.Split(resourceID, "/")
		if len(parts) == 2 {
			return "assumed-role", parts[0], parts[1]
		}
		return "assumed-role", resourceID, ""
	case service == "iam" && resourceType == "role":
		return "role", resourceID, ""
	case service == "iam" && resourceType == "user":
		return "user", resourceID, ""
	default:
		return "unknown", resourceID, ""
	}
}
