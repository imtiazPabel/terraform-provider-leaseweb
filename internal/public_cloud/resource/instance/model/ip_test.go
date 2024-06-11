package model

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/leaseweb/leaseweb-go-sdk/publicCloud"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_newIp(t *testing.T) {
	sdkDdos := publicCloud.NewNullableDdos(publicCloud.NewDdos())
	sdkDdos.Get().SetProtectionType("protection-type")

	sdkIp := publicCloud.NewIp()
	sdkIp.SetIp("ip")
	sdkIp.SetPrefixLength("prefix-length")
	sdkIp.SetVersion(46)
	sdkIp.SetNullRouted(true)
	sdkIp.SetMainIp(false)
	sdkIp.SetNetworkType("tralala")
	sdkIp.SetReverseLookup("reverse-lookup")
	sdkIp.Ddos = *sdkDdos

	ip, _ := newIp(context.TODO(), sdkIp)

	assert.Equal(t, "ip", ip.Ip.ValueString(), "ip should be set")
	assert.Equal(t, "prefix-length", ip.PrefixLength.ValueString(), "prefix-length should be set")
	assert.Equal(t, int64(46), ip.Version.ValueInt64(), "version should be set")
	assert.Equal(t, true, ip.NullRouted.ValueBool(), "nullRouted should be set")
	assert.Equal(t, false, ip.MainIp.ValueBool(), "mainIp should be set")
	assert.Equal(t, "tralala", ip.NetworkType.ValueString(), "networkType should be set")
	assert.Equal(t, "reverse-lookup", ip.ReverseLookup.ValueString(), "reverseLookup should be set")

	ddos := Ddos{}
	ip.Ddos.As(context.TODO(), &ddos, basetypes.ObjectAsOptions{})
	assert.Equal(t, "protection-type", ddos.ProtectionType.ValueString(), "ddos should be set")
}

func TestIp_attributeTypes(t *testing.T) {
	ip, _ := newIp(context.TODO(), publicCloud.NewIp())

	_, diags := types.ObjectValueFrom(context.TODO(), ip.attributeTypes(), ip)

	assert.Nil(t, diags, "attributes should be correct")
}
