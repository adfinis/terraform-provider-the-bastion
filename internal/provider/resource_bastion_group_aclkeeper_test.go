// Copyright (c) Adfinis
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/adfinis/terraform-provider-bastion/internal/provider/testutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccGroupACLKeeperResource(t *testing.T) {
	err := testutils.CreateAccount("testuser1")
	if err != nil {
		t.Errorf("Unable to create test account: %s", err)
	}

	t.Cleanup(func() {
		err := testutils.DeleteAccount("testuser1")
		if err != nil {
			t.Errorf("Unable to delete test account: %s", err)
		}
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccGroupACLKeeperResourceConfig("testgrpacl1", "bastionadmin", "testuser1"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_group_aclkeeper.test",
						tfjsonpath.New("group"),
						knownvalue.StringExact("testgrpacl1"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_aclkeeper.test",
						tfjsonpath.New("account"),
						knownvalue.StringExact("testuser1"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_aclkeeper.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("testgrpacl1:testuser1"),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "bastion_group_aclkeeper.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "testgrpacl1:testuser1",
			},
		},
	})
}

func TestAccGroupACLKeeperResource_Multiple(t *testing.T) {
	err := testutils.CreateAccounts("testuser2", "testuser3")
	if err != nil {
		t.Errorf("Unable to create test account: %s", err)
	}

	t.Cleanup(func() {
		err := testutils.DeleteAccounts("testuser2", "testuser3")
		if err != nil {
			t.Errorf("Unable to delete test account: %s", err)
		}
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create multiple ACL keepers
			{
				Config: testAccGroupACLKeeperResourceConfigMultiple("testgrpacl2", "bastionadmin"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bastion_group_aclkeeper.test1", "group", "testgrpacl2"),
					resource.TestCheckResourceAttr("bastion_group_aclkeeper.test1", "account", "testuser2"),
					resource.TestCheckResourceAttr("bastion_group_aclkeeper.test2", "group", "testgrpacl2"),
					resource.TestCheckResourceAttr("bastion_group_aclkeeper.test2", "account", "testuser3"),
				),
			},
		},
	})
}

func TestAccGroupACLKeeperResource_RequiresReplace(t *testing.T) {
	err := testutils.CreateAccounts("testuser4", "testuser5")
	if err != nil {
		t.Errorf("Unable to create test account: %s", err)
	}

	t.Cleanup(func() {
		err := testutils.DeleteAccounts("testuser4", "testuser5")
		if err != nil {
			t.Errorf("Unable to delete test account: %s", err)
		}
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupACLKeeperResourceConfig("testgrpacl3", "bastionadmin", "testuser4"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_group_aclkeeper.test",
						tfjsonpath.New("account"),
						knownvalue.StringExact("testuser4"),
					),
				},
			},
			// Change account should force replacement
			{
				Config: testAccGroupACLKeeperResourceConfig("testgrpacl3", "bastionadmin", "testuser5"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bastion_group_aclkeeper.test", "account", "testuser5"),
					resource.TestCheckResourceAttr("bastion_group_aclkeeper.test", "id", "testgrpacl3:testuser5"),
				),
			},
		},
	})
}

// testAccGroupACLKeeperResourceConfig generates the Terraform configuration for testing.
func testAccGroupACLKeeperResourceConfig(groupName, groupOwner, accountName string) string {
	config := providerConfig
	config += fmt.Sprintf(`
resource "bastion_group" "test" {
  group = %[1]q
  owner = %[2]q
}

resource "bastion_group_aclkeeper" "test" {
  group   = bastion_group.test.group
  account = %[3]q
}
`, groupName, groupOwner, accountName)

	return config
}

// testAccGroupACLKeeperResourceConfigMultiple generates config with multiple ACL keepers.
func testAccGroupACLKeeperResourceConfigMultiple(groupName, groupOwner string) string {
	config := providerConfig
	config += fmt.Sprintf(`
resource "bastion_group" "test" {
  group = %[1]q
  owner = %[2]q
}

resource "bastion_group_aclkeeper" "test1" {
  group   = bastion_group.test.group
  account = "testuser2"
}

resource "bastion_group_aclkeeper" "test2" {
  group   = bastion_group.test.group
  account = "testuser3"
}
`, groupName, groupOwner)

	return config
}
