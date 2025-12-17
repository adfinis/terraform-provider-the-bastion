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

func TestAccAccountCommandResource(t *testing.T) {
	err := testutils.CreateAccount("testcmduser1")
	if err != nil {
		t.Errorf("Unable to create test account: %s", err)
	}

	t.Cleanup(func() {
		err := testutils.DeleteAccount("testcmduser1")
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
				Config: testAccAccountCommandResourceConfig("testcmduser1", "selfAddPersonalAccess"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_account_command.test",
						tfjsonpath.New("account"),
						knownvalue.StringExact("testcmduser1"),
					),
					statecheck.ExpectKnownValue(
						"bastion_account_command.test",
						tfjsonpath.New("command"),
						knownvalue.StringExact("selfAddPersonalAccess"),
					),
					statecheck.ExpectKnownValue(
						"bastion_account_command.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("testcmduser1:selfAddPersonalAccess"),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "bastion_account_command.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "testcmduser1:selfAddPersonalAccess",
			},
		},
	})
}

func TestAccAccountCommandResource_Multiple(t *testing.T) {
	err := testutils.CreateAccount("testcmduser2")
	if err != nil {
		t.Errorf("Unable to create test account: %s", err)
	}

	t.Cleanup(func() {
		err := testutils.DeleteAccount("testcmduser2")
		if err != nil {
			t.Errorf("Unable to delete test account: %s", err)
		}
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create multiple command grants
			{
				Config: testAccAccountCommandResourceConfigMultiple("testcmduser2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bastion_account_command.test1", "account", "testcmduser2"),
					resource.TestCheckResourceAttr("bastion_account_command.test1", "command", "selfAddPersonalAccess"),
					resource.TestCheckResourceAttr("bastion_account_command.test2", "account", "testcmduser2"),
					resource.TestCheckResourceAttr("bastion_account_command.test2", "command", "selfDelPersonalAccess"),
				),
			},
		},
	})
}

func TestAccAccountCommandResource_RequiresReplace(t *testing.T) {
	err := testutils.CreateAccount("testcmduser3")
	if err != nil {
		t.Errorf("Unable to create test account: %s", err)
	}

	t.Cleanup(func() {
		err := testutils.DeleteAccount("testcmduser3")
		if err != nil {
			t.Errorf("Unable to delete test account: %s", err)
		}
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountCommandResourceConfig("testcmduser3", "selfAddPersonalAccess"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_account_command.test",
						tfjsonpath.New("command"),
						knownvalue.StringExact("selfAddPersonalAccess"),
					),
				},
			},
			// Change command should force replacement
			{
				Config: testAccAccountCommandResourceConfig("testcmduser3", "selfDelPersonalAccess"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bastion_account_command.test", "command", "selfDelPersonalAccess"),
					resource.TestCheckResourceAttr("bastion_account_command.test", "id", "testcmduser3:selfDelPersonalAccess"),
				),
			},
		},
	})
}

// testAccAccountCommandResourceConfig generates the Terraform configuration for testing.
func testAccAccountCommandResourceConfig(accountName, commandName string) string {
	config := providerConfig
	config += fmt.Sprintf(`
resource "bastion_account_command" "test" {
  account = %[1]q
  command = %[2]q
}
`, accountName, commandName)

	return config
}

// testAccAccountCommandResourceConfigMultiple generates config with multiple command grants.
func testAccAccountCommandResourceConfigMultiple(accountName string) string {
	config := providerConfig
	config += fmt.Sprintf(`
resource "bastion_account_command" "test1" {
  account = %[1]q
  command = "selfAddPersonalAccess"
}

resource "bastion_account_command" "test2" {
  account = %[1]q
  command = "selfDelPersonalAccess"
}
`, accountName)

	return config
}

func TestAccAccountCommandResource_Auditor(t *testing.T) {
	err := testutils.CreateAccount("testcmduser4")
	if err != nil {
		t.Errorf("Unable to create test account: %s", err)
	}

	t.Cleanup(func() {
		err := testutils.DeleteAccount("testcmduser4")
		if err != nil {
			t.Errorf("Unable to delete test account: %s", err)
		}
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing for auditor command
			{
				Config: testAccAccountCommandResourceConfig("testcmduser4", "auditor"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_account_command.test",
						tfjsonpath.New("account"),
						knownvalue.StringExact("testcmduser4"),
					),
					statecheck.ExpectKnownValue(
						"bastion_account_command.test",
						tfjsonpath.New("command"),
						knownvalue.StringExact("auditor"),
					),
					statecheck.ExpectKnownValue(
						"bastion_account_command.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("testcmduser4:auditor"),
					),
				},
			},
			// ImportState testing for auditor command
			{
				ResourceName:      "bastion_account_command.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "testcmduser4:auditor",
			},
		},
	})
}
