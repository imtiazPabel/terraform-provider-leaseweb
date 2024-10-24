package provider

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/leaseweb/leaseweb-go-sdk/dedicatedServer"
	"github.com/leaseweb/terraform-provider-leaseweb/internal/provider/client"
)

var (
	_ resource.Resource                = &dedicatedServerResource{}
	_ resource.ResourceWithConfigure   = &dedicatedServerResource{}
	_ resource.ResourceWithImportState = &dedicatedServerResource{}
)

type dedicatedServerResource struct {
	// TODO: Refactor this part, apiKey shouldn't be here.
	apiKey string
	client dedicatedServer.DedicatedServerAPI
}

type dedicatedServerResourceData struct {
	ID                           types.String `tfsdk:"id"`
	Reference                    types.String `tfsdk:"reference"`
	ReverseLookup                types.String `tfsdk:"reverse_lookup"`
	DHCPLease                    types.String `tfsdk:"dhcp_lease"`
	PoweredOn                    types.Bool   `tfsdk:"powered_on"`
	PublicNetworkInterfaceOpened types.Bool   `tfsdk:"public_network_interface_opened"`
	PublicIPNullRouted           types.Bool   `tfsdk:"public_ip_null_routed"`
	PublicIP                     types.String `tfsdk:"public_ip"`
	RemoteManagementIP           types.String `tfsdk:"remote_management_ip"`
	InternalMAC                  types.String `tfsdk:"internal_mac"`
	Location                     types.Object `tfsdk:"location"`
}

type dedicatedServerLocationResourceData struct {
	Rack  types.String `tfsdk:"rack"`
	Site  types.String `tfsdk:"site"`
	Suite types.String `tfsdk:"suite"`
	Unit  types.String `tfsdk:"unit"`
}

func (l dedicatedServerLocationResourceData) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"rack":  types.StringType,
		"site":  types.StringType,
		"suite": types.StringType,
		"unit":  types.StringType,
	}
}

func NewDedicatedServerResource() resource.Resource {
	return &dedicatedServerResource{}
}

func (d *dedicatedServerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dedicated_server"
}

func (d *dedicatedServerResource) authContext(ctx context.Context) context.Context {
	return context.WithValue(
		ctx,
		dedicatedServer.ContextAPIKeys,
		map[string]dedicatedServer.APIKey{
			"X-LSW-Auth": {Key: d.apiKey, Prefix: ""},
		},
	)
}

func (d *dedicatedServerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	configuration := dedicatedServer.NewConfiguration()

	// TODO: Refactor this part, ProviderData can be managed directly, not within client.
	coreClient, ok := req.ProviderData.(client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf(
				"Expected client.Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)
		return
	}
	d.apiKey = coreClient.ProviderData.ApiKey
	if coreClient.ProviderData.Host != nil {
		configuration.Host = *coreClient.ProviderData.Host
	}
	if coreClient.ProviderData.Scheme != nil {
		configuration.Scheme = *coreClient.ProviderData.Scheme
	}

	apiClient := dedicatedServer.NewAPIClient(configuration)
	d.client = apiClient.DedicatedServerAPI
}

func (d *dedicatedServerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the server.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"reference": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Reference of server.",
				Validators: []validator.String{
					stringvalidator.LengthAtMost(100),
				},
			},
			"reverse_lookup": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The reverse lookup associated with the dedicated server public IP.",
			},
			"dhcp_lease": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The URL of PXE boot the dedicated server is booting from.",
			},
			"powered_on": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the dedicated server is powered on or not.",
			},
			"public_network_interface_opened": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the public network interface of the dedicated server is opened or not.",
			},
			"public_ip_null_routed": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the public IP of the dedicated server is null routed or not.",
			},
			"public_ip": schema.StringAttribute{
				Computed:    true,
				Description: "The public IP of the dedicated server.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"remote_management_ip": schema.StringAttribute{
				Computed:    true,
				Description: "The remote management IP of the dedicated server.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"internal_mac": schema.StringAttribute{
				Computed:    true,
				Description: "The MAC address of the interface connected to internal private network.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"location": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"rack": schema.StringAttribute{
						Computed:    true,
						Description: "the location rack",
					},
					"site": schema.StringAttribute{
						Computed:    true,
						Description: "the location site",
					},
					"suite": schema.StringAttribute{
						Computed:    true,
						Description: "the location suite",
					},
					"unit": schema.StringAttribute{
						Computed:    true,
						Description: "the location unit",
					},
				},
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (d *dedicatedServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data dedicatedServerResourceData
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	dedicatedServer, err := d.getServer(ctx, data.ID.ValueString())
	if err != nil {
		summary := "Reading dedicated server"
		resp.Diagnostics.AddError(summary, NewError(nil, err).Error())
		tflog.Error(ctx, fmt.Sprintf("%s %s", summary, NewError(nil, err).Error()))
		return
	}

	diags = resp.State.Set(ctx, &dedicatedServer)
	resp.Diagnostics.Append(diags...)
}

func (d *dedicatedServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	resource.ImportStatePassthroughID(
		ctx,
		path.Root("id"),
		req,
		resp,
	)

	dedicatedServer, err := d.getServer(ctx, req.ID)
	if err != nil {
		summary := "Importing dedicated server"
		resp.Diagnostics.AddError(summary, NewError(nil, err).Error())
		tflog.Error(ctx, fmt.Sprintf("%s %s", summary, NewError(nil, err).Error()))
		return
	}

	diags := resp.State.Set(ctx, dedicatedServer)
	resp.Diagnostics.Append(diags...)
}

func (d *dedicatedServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dedicatedServerResourceData
	planDiags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(planDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state dedicatedServerResourceData
	stateDiags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(stateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Updating reference
	if !plan.Reference.IsNull() && !plan.Reference.IsUnknown() {
		ropts := dedicatedServer.NewUpdateServerReferenceOpts(plan.Reference.ValueString())
		response, err := d.client.UpdateServerReference(d.authContext(ctx), state.ID.ValueString()).UpdateServerReferenceOpts(*ropts).Execute()
		if err != nil {
			summary := fmt.Sprintf("Error updating dedicated server reference with id: %q", plan.ID.ValueString())
			resp.Diagnostics.AddError(summary, NewError(response, err).Error())
			tflog.Error(ctx, fmt.Sprintf("%s %s", summary, NewError(response, err).Error()))
			return
		}
		state.Reference = plan.Reference
	}

	// Updating Power status
	if !plan.PoweredOn.IsNull() && !plan.PoweredOn.IsUnknown() {
		if plan.PoweredOn.ValueBool() {
			request := d.client.PowerServerOn(d.authContext(ctx), state.ID.ValueString())
			response, err := request.Execute()
			if err != nil {
				summary := fmt.Sprintf("Error powering on for dedicated server: %q", state.ID.ValueString())
				resp.Diagnostics.AddError(summary, NewError(response, err).Error())
				tflog.Error(ctx, fmt.Sprintf("%s %s", summary, NewError(response, err).Error()))
				return
			}
		} else {
			request := d.client.PowerServerOff(d.authContext(ctx), state.ID.ValueString())
			response, err := request.Execute()
			if err != nil {
				summary := fmt.Sprintf("Error powering off for dedicated server: %q", state.ID.ValueString())
				resp.Diagnostics.AddError(summary, NewError(response, err).Error())
				tflog.Error(ctx, fmt.Sprintf("%s %s", summary, NewError(response, err).Error()))
				return
			}
		}
		state.PoweredOn = plan.PoweredOn
	}

	// Updateing Reverse Lookup
	isPublicIPExists := !state.PublicIP.IsNull() && !state.PublicIP.IsUnknown() && state.PublicIP.ValueString() != ""
	if !plan.ReverseLookup.IsNull() && !plan.ReverseLookup.IsUnknown() && isPublicIPExists {
		iopts := dedicatedServer.NewUpdateIpProfileOpts()
		iopts.ReverseLookup = plan.ReverseLookup.ValueStringPointer()
		_, response, err := d.client.UpdateIpProfile(d.authContext(ctx), state.ID.ValueString(), state.PublicIP.ValueString()).UpdateIpProfileOpts(*iopts).Execute()
		if err != nil {
			summary := fmt.Sprintf("Error updating dedicated server reverse lookup with id: %q", state.ID.ValueString())
			resp.Diagnostics.AddError(summary, NewError(response, err).Error())
			tflog.Error(ctx, fmt.Sprintf("%s %s", summary, NewError(response, err).Error()))
			return
		}
		state.ReverseLookup = plan.ReverseLookup
	}

	// Updating an IP null routing
	if !plan.PublicIPNullRouted.IsNull() && !plan.PublicIPNullRouted.IsUnknown() && plan.PublicIPNullRouted != state.PublicIPNullRouted && isPublicIPExists {
		if plan.PublicIPNullRouted.ValueBool() {
			_, response, err := d.client.NullIpRoute(d.authContext(ctx), state.ID.ValueString(), state.PublicIP.ValueString()).Execute()
			if err != nil {
				summary := fmt.Sprintf("Error null routing an IP for dedicated server: %q and IP: %q", state.ID.ValueString(), state.PublicIP.ValueString())
				resp.Diagnostics.AddError(summary, NewError(response, err).Error())
				tflog.Error(ctx, fmt.Sprintf("%s %s", summary, NewError(response, err).Error()))
				return
			}
		} else {
			_, response, err := d.client.RemoveNullIpRoute(d.authContext(ctx), state.ID.ValueString(), state.PublicIP.ValueString()).Execute()
			if err != nil {
				summary := fmt.Sprintf("Error remove null routing an IP for dedicated server: %q and IP: %q", state.ID.ValueString(), state.PublicIP.ValueString())
				resp.Diagnostics.AddError(summary, NewError(response, err).Error())
				tflog.Error(ctx, fmt.Sprintf("%s %s", summary, NewError(response, err).Error()))
				return
			}
		}
		state.PublicIPNullRouted = plan.PublicIPNullRouted
	}

	// Updating dhcp lease
	if !plan.DHCPLease.IsNull() && !plan.DHCPLease.IsUnknown() {
		if plan.DHCPLease.ValueString() != "" {
			opts := dedicatedServer.NewCreateServerDhcpReservationOpts(plan.DHCPLease.ValueString())
			response, err := d.client.CreateServerDhcpReservation(d.authContext(ctx), state.ID.ValueString()).CreateServerDhcpReservationOpts(*opts).Execute()
			if err != nil {
				summary := fmt.Sprintf("Error creating a DHCP reservation for dedicated server: %q", state.ID.ValueString())
				resp.Diagnostics.AddError(summary, NewError(response, err).Error())
				tflog.Error(ctx, fmt.Sprintf("%s %s", summary, NewError(response, err).Error()))
				return
			}
		} else {
			response, err := d.client.DeleteServerDhcpReservation(d.authContext(ctx), state.ID.ValueString()).Execute()
			if err != nil {
				summary := fmt.Sprintf("Error deleting DHCP reservation for dedicated server: %q", state.ID.ValueString())
				resp.Diagnostics.AddError(summary, NewError(response, err).Error())
				tflog.Error(ctx, fmt.Sprintf("%s %s", summary, NewError(response, err).Error()))
				return
			}
		}
		state.DHCPLease = plan.DHCPLease
	}

	// Updating network interface status
	if !plan.PublicIPNullRouted.IsNull() && !plan.PublicIPNullRouted.IsUnknown() && plan.PublicNetworkInterfaceOpened != state.PublicNetworkInterfaceOpened {
		if plan.PublicNetworkInterfaceOpened.ValueBool() {
			response, err := d.client.OpenNetworkInterface(d.authContext(ctx), state.ID.ValueString(), dedicatedServer.NETWORKTYPEURL_PUBLIC).Execute()
			if err != nil {
				summary := fmt.Sprintf("Error opening public network interface for dedicated server: %q", state.ID.ValueString())
				resp.Diagnostics.AddError(summary, NewError(response, err).Error())
				tflog.Error(ctx, fmt.Sprintf("%s %s", summary, NewError(response, err).Error()))
				return
			}
		} else {
			response, err := d.client.CloseNetworkInterface(d.authContext(ctx), state.ID.ValueString(), dedicatedServer.NETWORKTYPEURL_PUBLIC).Execute()
			if err != nil {
				summary := fmt.Sprintf("Error closing public network interface for dedicated server: %q", state.ID.ValueString())
				resp.Diagnostics.AddError(summary, NewError(response, err).Error())
				tflog.Error(ctx, fmt.Sprintf("%s %s", summary, NewError(response, err).Error()))
				return
			}
		}
		state.PublicNetworkInterfaceOpened = plan.PublicNetworkInterfaceOpened
	}

	stateDiags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(stateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *dedicatedServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	panic("unimplemented")
}

func (d *dedicatedServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	panic("unimplemented")
}

func (d *dedicatedServerResource) getServer(ctx context.Context, serverID string) (*dedicatedServerResourceData, error) {

	// Getting server info
	serverResult, serverResponse, err := d.client.GetServer(d.authContext(ctx), serverID).Execute()
	if err != nil {
		return nil, fmt.Errorf("error reading dedicated server with id: %q - %s", serverID, NewError(serverResponse, err).Error())
	}

	var publicIP string
	var publicIPNullRouted bool
	if networkInterfaces, ok := serverResult.GetNetworkInterfacesOk(); ok {
		if publicNetworkInterface, ok := networkInterfaces.GetPublicOk(); ok {
			publicIPPart := strings.Split(publicNetworkInterface.GetIp(), "/")
			ip := net.ParseIP(publicIPPart[0])
			if ip != nil {
				publicIP = ip.String()
			}
			publicIPNullRouted = publicNetworkInterface.GetNullRouted()
		}
	}

	var reference string
	if contract, ok := serverResult.GetContractOk(); ok {
		reference = contract.GetReference()
	}

	var internalMAC string
	if networkInterfaces, ok := serverResult.GetNetworkInterfacesOk(); ok {
		if internalNetworkInterface, ok := networkInterfaces.GetInternalOk(); ok {
			internalMAC = internalNetworkInterface.GetMac()
		}
	}

	var remoteManagementIP string
	if networkInterfaces, ok := serverResult.GetNetworkInterfacesOk(); ok {
		if remoteNetworkInterface, ok := networkInterfaces.GetRemoteManagementOk(); ok {
			remoteManagementIPPart := strings.Split(remoteNetworkInterface.GetIp(), "/")
			ip := net.ParseIP(remoteManagementIPPart[0])
			if ip != nil {
				remoteManagementIP = ip.String()
			}
		}
	}

	serverLocation := serverResult.GetLocation()
	l := dedicatedServerLocationResourceData{
		Rack:  types.StringValue(serverLocation.GetRack()),
		Site:  types.StringValue(serverLocation.GetSite()),
		Suite: types.StringValue(serverLocation.GetSuite()),
		Unit:  types.StringValue(serverLocation.GetUnit()),
	}
	location, digs := types.ObjectValueFrom(ctx, l.AttributeTypes(), l)
	if digs.HasError() {
		return nil, fmt.Errorf("error reading dedicated server location with id: %q", serverID)
	}

	// Getting server power info
	powerResult, powerResponse, err := d.client.GetServerPowerStatus(d.authContext(ctx), serverID).Execute()
	if err != nil {
		return nil, fmt.Errorf("error reading dedicated server power status with id: %q - %s", serverID, NewError(powerResponse, err).Error())
	}
	pdu := powerResult.GetPdu()
	ipmi := powerResult.GetIpmi()
	poweredOn := pdu.GetStatus() != "off" && ipmi.GetStatus() != "off"

	// Getting server public network interface info
	var publicNetworkOpened bool
	networkRequest := d.client.GetNetworkInterface(d.authContext(ctx), serverID, dedicatedServer.NETWORKTYPEURL_PUBLIC)
	networkResult, networkResponse, err := networkRequest.Execute()
	if err != nil && networkResponse.StatusCode != http.StatusNotFound {
		return nil, fmt.Errorf("error reading dedicated server network interface with id: %q - %s", serverID, NewError(networkResponse, err).Error())
	} else {
		if _, ok := networkResult.GetStatusOk(); ok {
			publicNetworkOpened = networkResult.GetStatus() == "open"
		}
	}

	// Getting server DHCP info
	dhcpResult, dhcpResponse, err := d.client.GetServerDhcpReservationList(d.authContext(ctx), serverID).Execute()
	if err != nil {
		return nil, fmt.Errorf("error reading dedicated server DHCP with id: %q - %s", serverID, NewError(dhcpResponse, err).Error())
	}
	var dhcpLease string
	if len(dhcpResult.GetLeases()) != 0 {
		leases := dhcpResult.GetLeases()
		dhcpLease = leases[0].GetBootfile()
	}

	// Getting server public IP info
	var reverseLookup string
	if publicIP != "" {
		ipResult, ipResponse, err := d.client.GetServerIp(d.authContext(ctx), serverID, publicIP).Execute()
		if err != nil {
			return nil, fmt.Errorf("error reading dedicated server IP details with id: %q - %s", serverID, NewError(ipResponse, err).Error())
		}
		reverseLookup = ipResult.GetReverseLookup()
	}

	dedicatedServer := dedicatedServerResourceData{
		ID:                           types.StringValue(serverResult.GetId()),
		Reference:                    types.StringValue(reference),
		ReverseLookup:                types.StringValue(reverseLookup),
		DHCPLease:                    types.StringValue(dhcpLease),
		PoweredOn:                    types.BoolValue(poweredOn),
		PublicNetworkInterfaceOpened: types.BoolValue(publicNetworkOpened),
		PublicIPNullRouted:           types.BoolValue(publicIPNullRouted),
		PublicIP:                     types.StringValue(publicIP),
		RemoteManagementIP:           types.StringValue(remoteManagementIP),
		InternalMAC:                  types.StringValue(internalMAC),
		Location:                     location,
	}

	return &dedicatedServer, nil
}
