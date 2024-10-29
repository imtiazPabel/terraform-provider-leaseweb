package publiccloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"
)

func Test_contractTermValidator_ValidateObject(t *testing.T) {
	t.Run(
		"does not set error if contract term is correct",
		func(t *testing.T) {
			contract := contractResourceModel{}
			configValue, _ := types.ObjectValueFrom(
				context.TODO(),
				contract.AttributeTypes(),
				contract,
			)

			request := validator.ObjectRequest{
				ConfigValue: configValue,
			}

			response := validator.ObjectResponse{}

			contractTermValidator := contractTermValidator{}
			contractTermValidator.ValidateObject(context.TODO(), request, &response)

			assert.Len(t, response.Diagnostics.Errors(), 0)
		},
	)

	t.Run(
		"returns expected error if contract term cannot be 0",
		func(t *testing.T) {
			contract := contractResourceModel{
				Type: basetypes.NewStringValue("MONTHLY"),
				Term: basetypes.NewInt64Value(0),
			}
			configValue, _ := types.ObjectValueFrom(
				context.TODO(),
				contract.AttributeTypes(),
				contract,
			)

			request := validator.ObjectRequest{
				ConfigValue: configValue,
			}

			response := validator.ObjectResponse{}

			contractTermValidator := contractTermValidator{}
			contractTermValidator.ValidateObject(context.TODO(), request, &response)

			assert.Len(t, response.Diagnostics.Errors(), 1)
			assert.Contains(
				t,
				response.Diagnostics.Errors()[0].Detail(),
				"MONTHLY",
			)
		},
	)

	t.Run(
		"returns expected error if contract term must be 0",
		func(t *testing.T) {
			contract := contractResourceModel{
				Type: basetypes.NewStringValue("HOURLY"),
				Term: basetypes.NewInt64Value(3),
			}
			configValue, _ := types.ObjectValueFrom(
				context.TODO(),
				contract.AttributeTypes(),
				contract,
			)

			request := validator.ObjectRequest{
				ConfigValue: configValue,
			}

			response := validator.ObjectResponse{}

			contractTermValidator := contractTermValidator{}
			contractTermValidator.ValidateObject(context.TODO(), request, &response)

			assert.Len(t, response.Diagnostics.Errors(), 1)
			assert.Contains(
				t,
				response.Diagnostics.Errors()[0].Detail(),
				"HOURLY",
			)
		},
	)
}

func Test_instanceTerminationValidator_ValidateObject(t *testing.T) {
	t.Run("ConfigValue populate errors bubble up", func(t *testing.T) {
		request := validator.ObjectRequest{}
		response := validator.ObjectResponse{}

		instanceTerminationValidator := instanceTerminationValidator{}
		instanceTerminationValidator.ValidateObject(context.TODO(), request, &response)

		assert.True(t, response.Diagnostics.HasError())
		assert.Contains(
			t,
			response.Diagnostics[0].Summary(),
			"Value Conversion Error",
		)
	})

	t.Run(
		"does not set a diagnostics error if instance is allowed to be terminated",
		func(t *testing.T) {
			instance := generateInstanceModelForValidator()
			instanceObject, _ := basetypes.NewObjectValueFrom(
				context.TODO(),
				instance.AttributeTypes(),
				instance,
			)
			request := validator.ObjectRequest{ConfigValue: instanceObject}
			response := validator.ObjectResponse{}

			instanceTerminationValidator := instanceTerminationValidator{}
			instanceTerminationValidator.ValidateObject(context.TODO(), request, &response)

			assert.False(t, response.Diagnostics.HasError())
		},
	)

	t.Run(
		"sets a diagnostics error if instance is not allowed to be terminated",
		func(t *testing.T) {
			instance := generateInstanceModelForValidator()
			instance.State = basetypes.NewStringValue("DESTROYED")
			instanceObject, _ := basetypes.NewObjectValueFrom(
				context.TODO(),
				instance.AttributeTypes(),
				instance,
			)
			request := validator.ObjectRequest{ConfigValue: instanceObject}
			response := validator.ObjectResponse{}

			instanceTerminationValidator := instanceTerminationValidator{}
			instanceTerminationValidator.ValidateObject(context.TODO(), request, &response)

			assert.True(t, response.Diagnostics.HasError())
			assert.Contains(t, response.Diagnostics[0].Detail(), "DESTROYED")
		},
	)
}

func generateInstanceModelForValidator() instanceResourceModel {
	contract := contractResourceModel{}
	contractObject, _ := types.ObjectValueFrom(
		context.TODO(),
		contract.AttributeTypes(),
		contract,
	)

	return instanceResourceModel{
		ID:        basetypes.NewStringUnknown(),
		Region:    basetypes.NewStringUnknown(),
		Reference: basetypes.NewStringUnknown(),
		Image: basetypes.NewObjectUnknown(
			imageResourceModel{}.AttributeTypes(),
		),
		State:               basetypes.NewStringUnknown(),
		Type:                basetypes.NewStringUnknown(),
		RootDiskSize:        basetypes.NewInt64Unknown(),
		RootDiskStorageType: basetypes.NewStringUnknown(),
		IPs: basetypes.NewListUnknown(
			types.ObjectType{
				AttrTypes: iPResourceModel{}.AttributeTypes(),
			},
		),
		Contract:    contractObject,
		MarketAppID: basetypes.NewStringUnknown(),
	}
}
