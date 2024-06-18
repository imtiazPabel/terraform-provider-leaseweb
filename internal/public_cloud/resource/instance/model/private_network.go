package model

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/leaseweb/leaseweb-go-sdk/publicCloud"
)

type PrivateNetwork struct {
	Id     types.String `tfsdk:"id"`
	Status types.String `tfsdk:"status"`
	Subnet types.String `tfsdk:"subnet"`
}

func (p PrivateNetwork) attributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":     types.StringType,
		"status": types.StringType,
		"subnet": types.StringType,
	}
}

func newPrivateNetwork(sdkPrivateNetwork publicCloud.PrivateNetwork) PrivateNetwork {
	return PrivateNetwork{
		Id:     basetypes.NewStringValue(sdkPrivateNetwork.GetPrivateNetworkId()),
		Status: basetypes.NewStringValue(sdkPrivateNetwork.GetStatus()),
		Subnet: basetypes.NewStringValue(sdkPrivateNetwork.GetSubnet()),
	}
}
