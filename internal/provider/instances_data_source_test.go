package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccInstancesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `data "leaseweb_instances" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.leaseweb_instances.test",
						"instances.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						"data.leaseweb_instances.test",
						"instances.0.id",
						"ace712e9-a166-47f1-9065-4af0f7e7fce1",
					),
				),
			},
		},
	})
}
