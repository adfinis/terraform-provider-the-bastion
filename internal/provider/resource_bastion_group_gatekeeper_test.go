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

func TestAccGroupGatekeeperResource(t *testing.T) {
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
				Config: testAccGroupGatekeeperResourceConfig("testgrpgate1", "bastionadmin", "testuser1"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_group_gatekeeper.test",
						tfjsonpath.New("group"),
						knownvalue.StringExact("testgrpgate1"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_gatekeeper.test",
						tfjsonpath.New("account"),
						knownvalue.StringExact("testuser1"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_gatekeeper.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("testgrpgate1:testuser1"),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "bastion_group_gatekeeper.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "testgrpgate1:testuser1",
			},
		},
	})
}

func TestAccGroupGatekeeperResource_Multiple(t *testing.T) {
	err := testutils.CreateAccounts("testuser2", "testuser3")
	if err != nil {
		t.Errorf("Unable to create test accounts: %s", err)
	}

	t.Cleanup(func() {
		err := testutils.DeleteAccounts("testuser2", "testuser3")
		if err != nil {
			t.Errorf("Unable to delete test accounts: %s", err)
		}
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create multiple gatekeepers
			{
				Config: testAccGroupGatekeeperResourceConfigMultiple("testgrpgate2", "bastionadmin"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bastion_group_gatekeeper.test1", "group", "testgrpgate2"),
					resource.TestCheckResourceAttr("bastion_group_gatekeeper.test1", "account", "testuser2"),
					resource.TestCheckResourceAttr("bastion_group_gatekeeper.test2", "group", "testgrpgate2"),
					resource.TestCheckResourceAttr("bastion_group_gatekeeper.test2", "account", "testuser3"),
				),
			},
		},
	})
}

func TestAccGroupGatekeeperResource_RequiresReplace(t *testing.T) {
	err := testutils.CreateAccounts("testuser4", "testuser5")
	if err != nil {
		t.Errorf("Unable to create test accounts: %s", err)
	}

	t.Cleanup(func() {
		err := testutils.DeleteAccounts("testuser4", "testuser5")
		if err != nil {
			t.Errorf("Unable to delete test accounts: %s", err)
		}
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupGatekeeperResourceConfig("testgrpgate3", "bastionadmin", "testuser4"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_group_gatekeeper.test",
						tfjsonpath.New("account"),
						knownvalue.StringExact("testuser4"),
					),
				},
			},
			// Change account should force replacement
			{
				Config: testAccGroupGatekeeperResourceConfig("testgrpgate3", "bastionadmin", "testuser5"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bastion_group_gatekeeper.test", "account", "testuser5"),
					resource.TestCheckResourceAttr("bastion_group_gatekeeper.test", "id", "testgrpgate3:testuser5"),
				),
			},
		},
	})
}

// testAccGroupGatekeeperResourceConfig generates the Terraform configuration for testing.
func testAccGroupGatekeeperResourceConfig(groupName, groupOwner, accountName string) string {
	config := providerConfig
	config += fmt.Sprintf(`
resource "bastion_group" "test" {
  group = %[1]q
  owner = %[2]q
}

resource "bastion_group_gatekeeper" "test" {
  group   = bastion_group.test.group
  account = %[3]q
}
`, groupName, groupOwner, accountName)

	return config
}

// testAccGroupGatekeeperResourceConfigMultiple generates config with multiple gatekeepers.
func testAccGroupGatekeeperResourceConfigMultiple(groupName, groupOwner string) string {
	config := providerConfig
	config += fmt.Sprintf(`
resource "bastion_group" "test" {
  group = %[1]q
  owner = %[2]q
}

resource "bastion_group_gatekeeper" "test1" {
  group   = bastion_group.test.group
  account = "testuser2"
}

resource "bastion_group_gatekeeper" "test2" {
  group   = bastion_group.test.group
  account = "testuser3"
}
`, groupName, groupOwner)

	return config
}
