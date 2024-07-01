package model

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/leaseweb/leaseweb-go-sdk/publicCloud"
	"github.com/stretchr/testify/assert"
)

func Test_newDdos(t *testing.T) {
	sdkDdos := publicCloud.NewNullableDdos(publicCloud.NewDdos(
		"detectionProfile",
		"protectionType",
	))

	got := newDdos(sdkDdos.Get())

	assert.Equal(
		t,
		"detectionProfile",
		got.DetectionProfile.ValueString(),
		"detectionProfile should be set",
	)
	assert.Equal(
		t,
		"protectionType",
		got.ProtectionType.ValueString(),
		"protectionType should be set",
	)
}

func TestDdos_attributeTypes(t *testing.T) {
	_, diags := types.ObjectValueFrom(
		context.TODO(),
		Ddos{}.AttributeTypes(),
		Ddos{},
	)

	assert.Nil(t, diags, "attributes should be correct")
}
