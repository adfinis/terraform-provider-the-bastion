// Copyright (c) Adfinis
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGroupDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create a group first, then read it with the data source
			{
				Config: testAccGroupDataSourceConfig("testgroup-ds", "bastionadmin"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify resource attributes
					resource.TestCheckResourceAttr("bastion_group.test", "group", "testgroup-ds"),
					resource.TestCheckResourceAttr("bastion_group.test", "owner", "bastionadmin"),
					// Verify data source attributes
					resource.TestCheckResourceAttr("data.bastion_group.test", "group", "testgroup-ds"),
					resource.TestCheckResourceAttrSet("data.bastion_group.test", "owners.#"),
					resource.TestCheckResourceAttrSet("data.bastion_group.test", "members.#"),
					resource.TestCheckResourceAttrSet("data.bastion_group.test", "gatekeepers.#"),
					resource.TestCheckResourceAttrSet("data.bastion_group.test", "aclkeepers.#"),
					// Verify owner is in the owners list
					resource.TestCheckResourceAttr("data.bastion_group.test", "owners.0", "bastionadmin"),
				),
			},
		},
	})
}

func testAccGroupDataSourceConfig(groupName, owner string) string {
	return providerConfig + fmt.Sprintf(`
resource "bastion_group" "test" {
  group = %[1]q
  owner = %[2]q
}

data "bastion_group" "test" {
  group = bastion_group.test.group
}
`, groupName, owner)
}
