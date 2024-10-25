package publiccloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/leaseweb/leaseweb-go-sdk/publicCloud"
	"github.com/stretchr/testify/assert"
)

func Test_adaptSdkLoadBalancerDetailsToResourceLoadBalancer(t *testing.T) {
	t.Run("required fields are set", func(t *testing.T) {
		loadBalancerDetails := publicCloud.LoadBalancerDetails{
			Id:        "id",
			Region:    "region",
			Type:      publicCloud.TYPENAME_C3_2XLARGE,
			Reference: *publicCloud.NewNullableString(nil),
			Contract: publicCloud.Contract{
				Type: publicCloud.CONTRACTTYPE_MONTHLY,
			},
		}

		got, err := adaptSdkLoadBalancerDetailsToResourceLoadBalancer(
			loadBalancerDetails,
			context.TODO(),
		)

		assert.NoError(t, err)
		assert.Equal(t, "id", got.ID.ValueString())
		assert.Equal(t, "region", got.Region.ValueString())
		assert.Equal(t, "lsw.c3.2xlarge", got.Type.ValueString())
		assert.Nil(t, got.Reference.ValueStringPointer())

		contract := resourceModelContract{}
		got.Contract.As(context.TODO(), &contract, basetypes.ObjectAsOptions{})
		assert.Equal(t, "MONTHLY", contract.Type.ValueString())
	})

	t.Run("optional fields are set", func(t *testing.T) {
		reference := "reference"

		loadBalancerDetails := publicCloud.LoadBalancerDetails{
			Id:        "id",
			Region:    "region",
			Type:      publicCloud.TYPENAME_C3_2XLARGE,
			Reference: *publicCloud.NewNullableString(&reference),
			Contract: publicCloud.Contract{
				Type: publicCloud.CONTRACTTYPE_MONTHLY,
			},
		}

		got, err := adaptSdkLoadBalancerDetailsToResourceLoadBalancer(
			loadBalancerDetails,
			context.TODO(),
		)

		assert.NoError(t, err)
		assert.Equal(t, "reference", got.Reference.ValueString())
	})

}

func Test_resourceModelLoadBalancer_GetLaunchLoadBalancerOpts(t *testing.T) {
	t.Run("required values are set", func(t *testing.T) {
		loadBalancer := generateLoadBalancerModel()
		loadBalancer.Reference = basetypes.NewStringPointerValue(nil)

		got, err := loadBalancer.GetLaunchLoadBalancerOpts(context.TODO())

		assert.NoError(t, err)
		assert.Equal(t, publicCloud.REGIONNAME_EU_WEST_3, got.Region)
		assert.Equal(t, publicCloud.TYPENAME_C3_2XLARGE, got.Type)
		assert.Equal(t, publicCloud.CONTRACTTYPE_MONTHLY, got.ContractType)
		assert.Equal(t, publicCloud.CONTRACTTERM__3, got.ContractTerm)
		assert.Equal(t, publicCloud.BILLINGFREQUENCY__1, got.BillingFrequency)

		reference, _ := got.GetReferenceOk()
		assert.Nil(t, reference)
	})

	t.Run("optional values are passed", func(t *testing.T) {
		reference := "reference"

		loadBalancer := generateLoadBalancerModel()
		loadBalancer.Reference = basetypes.NewStringPointerValue(&reference)

		got, err := loadBalancer.GetLaunchLoadBalancerOpts(context.TODO())

		assert.NoError(t, err)
		assert.Equal(t, "reference", *got.Reference)
	})

	t.Run(
		"returns error if invalid instanceType is passed",
		func(t *testing.T) {
			loadBalancer := generateLoadBalancerModel()
			loadBalancer.Type = basetypes.NewStringValue("tralala")

			_, err := loadBalancer.GetLaunchLoadBalancerOpts(context.TODO())

			assert.Error(t, err)
			assert.ErrorContains(t, err, "tralala")
		},
	)

	t.Run(
		"returns error if invalid contractType is passed",
		func(t *testing.T) {
			contractType := "tralala"
			loadBalancer := generateLoadBalancerModel()
			contract := GenerateContractObject(
				nil,
				nil,
				&contractType,
				nil,
			)
			loadBalancer.Contract = contract

			_, err := loadBalancer.GetLaunchLoadBalancerOpts(context.TODO())

			assert.Error(t, err)
			assert.ErrorContains(t, err, "tralala")
		},
	)

	t.Run(
		"returns error if invalid contractTerm is passed",
		func(t *testing.T) {
			contractTerm := 555
			loadBalancer := generateLoadBalancerModel()
			contract := GenerateContractObject(
				nil,
				&contractTerm,
				nil,
				nil,
			)
			loadBalancer.Contract = contract

			_, err := loadBalancer.GetLaunchLoadBalancerOpts(context.TODO())

			assert.Error(t, err)
			assert.ErrorContains(t, err, "555")
		},
	)

	t.Run(
		"returns error if invalid billingFrequency is passed",
		func(t *testing.T) {
			billingFrequency := 555
			loadBalancer := generateLoadBalancerModel()
			contract := GenerateContractObject(
				&billingFrequency,
				nil,
				nil,
				nil,
			)
			loadBalancer.Contract = contract

			_, err := loadBalancer.GetLaunchLoadBalancerOpts(context.TODO())

			assert.Error(t, err)
			assert.ErrorContains(t, err, "555")
		},
	)

	t.Run(
		"returns error if invalid region is passed",
		func(t *testing.T) {
			loadBalancer := generateLoadBalancerModel()
			loadBalancer.Region = basetypes.NewStringValue("tralala")

			_, err := loadBalancer.GetLaunchLoadBalancerOpts(context.TODO())

			assert.Error(t, err)
			assert.ErrorContains(t, err, "tralala")
		},
	)

	t.Run(
		"returns error if resourceModelContract resource is incorrect",
		func(t *testing.T) {
			loadBalancer := generateLoadBalancerModel()
			loadBalancer.Contract = basetypes.NewObjectNull(map[string]attr.Type{})

			_, err := loadBalancer.GetLaunchLoadBalancerOpts(context.TODO())

			assert.Error(t, err)
			assert.ErrorContains(t, err, ".resourceModelContract")
		},
	)
}

func Test_resourceModelLoadBalancer_GetUpdateLoadBalancerOpts(t *testing.T) {
	t.Run("optional values are set", func(t *testing.T) {
		reference := "reference"
		loadBalancerType := string(publicCloud.TYPENAME_C3_2XLARGE)

		loadBalancer := generateLoadBalancerModel()
		loadBalancer.Type = basetypes.NewStringPointerValue(&loadBalancerType)
		loadBalancer.Reference = basetypes.NewStringPointerValue(&reference)

		got, err := loadBalancer.GetUpdateLoadBalancerOpts()

		assert.NoError(t, err)
		assert.Equal(t, publicCloud.TYPENAME_C3_2XLARGE, *got.Type)
		assert.Equal(t, "reference", *got.Reference)
	})

	t.Run(
		"returns error if invalid instanceType is passed",
		func(t *testing.T) {
			loadBalancer := generateLoadBalancerModel()
			loadBalancer.Type = basetypes.NewStringValue("tralala")

			_, err := loadBalancer.GetUpdateLoadBalancerOpts()

			assert.Error(t, err)
			assert.ErrorContains(t, err, "tralala")
		},
	)
}

func generateLoadBalancerModel() resourceModelLoadBalancer {
	contract := GenerateContractObject(
		nil,
		nil,
		nil,
		nil,
	)

	return resourceModelLoadBalancer{
		ID:        basetypes.NewStringValue("305c0bd8-b157-4a9c-885a-e07df86a714f"),
		Region:    basetypes.NewStringValue(string(publicCloud.REGIONNAME_EU_WEST_3)),
		Type:      basetypes.NewStringValue(string(publicCloud.TYPENAME_C3_2XLARGE)),
		Reference: basetypes.NewStringPointerValue(nil),
		Contract:  contract,
	}
}
