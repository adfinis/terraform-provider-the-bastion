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

func TestAccGroupOwnerResource(t *testing.T) {
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
				Config: testAccGroupOwnerResourceConfig("testgrpowner1", "bastionadmin", "testuser1"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_group_owner.test",
						tfjsonpath.New("group"),
						knownvalue.StringExact("testgrpowner1"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_owner.test",
						tfjsonpath.New("account"),
						knownvalue.StringExact("testuser1"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_owner.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("testgrpowner1:testuser1"),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "bastion_group_owner.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "testgrpowner1:testuser1",
			},
		},
	})
}

func TestAccGroupOwnerResource_Multiple(t *testing.T) {
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
			// Create multiple owners
			{
				Config: testAccGroupOwnerResourceConfigMultiple("testgrpowner2", "bastionadmin"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bastion_group_owner.test1", "group", "testgrpowner2"),
					resource.TestCheckResourceAttr("bastion_group_owner.test1", "account", "testuser2"),
					resource.TestCheckResourceAttr("bastion_group_owner.test2", "group", "testgrpowner2"),
					resource.TestCheckResourceAttr("bastion_group_owner.test2", "account", "testuser3"),
				),
			},
		},
	})
}

func TestAccGroupOwnerResource_RequiresReplace(t *testing.T) {
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
				Config: testAccGroupOwnerResourceConfig("testgrpowner3", "bastionadmin", "testuser4"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_group_owner.test",
						tfjsonpath.New("account"),
						knownvalue.StringExact("testuser4"),
					),
				},
			},
			// Change account should force replacement
			{
				Config: testAccGroupOwnerResourceConfig("testgrpowner3", "bastionadmin", "testuser5"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bastion_group_owner.test", "account", "testuser5"),
					resource.TestCheckResourceAttr("bastion_group_owner.test", "id", "testgrpowner3:testuser5"),
				),
			},
		},
	})
}

// testAccGroupOwnerResourceConfig generates the Terraform configuration for testing.
func testAccGroupOwnerResourceConfig(groupName, groupOwner, accountName string) string {
	config := providerConfig
	config += fmt.Sprintf(`
resource "bastion_group" "test" {
  group = %[1]q
  owner = %[2]q
}

resource "bastion_group_owner" "test" {
  group   = bastion_group.test.group
  account = %[3]q
}
`, groupName, groupOwner, accountName)

	return config
}

// testAccGroupOwnerResourceConfigMultiple generates config with multiple owners.
func testAccGroupOwnerResourceConfigMultiple(groupName, groupOwner string) string {
	config := providerConfig
	config += fmt.Sprintf(`
resource "bastion_group" "test" {
  group = %[1]q
  owner = %[2]q
}

resource "bastion_group_owner" "test1" {
  group   = bastion_group.test.group
  account = "testuser2"
}

resource "bastion_group_owner" "test2" {
  group   = bastion_group.test.group
  account = "testuser3"
}
`, groupName, groupOwner)

	return config
}
