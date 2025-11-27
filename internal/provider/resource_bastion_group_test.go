// Copyright (c) Adfinis
// SPDX-License-Identifier: GPL-3.0-or-later

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccGroupResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupResourceConfig("testgrp1", "bastionadmin", ""),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_group.test",
						tfjsonpath.New("group"),
						knownvalue.StringExact("testgrp1"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group.test",
						tfjsonpath.New("owner"),
						knownvalue.StringExact("bastionadmin"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group.test",
						tfjsonpath.New("key_algo"),
						knownvalue.StringExact("ed25519"),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:                         "bastion_group.test",
				ImportStateVerifyIdentifierAttribute: "group",
				ImportStateId:                        "testgrp1",
				ImportState:                          true,
				ImportStateVerify:                    true,
				// owner and key_algo are only used during creation
				ImportStateVerifyIgnore: []string{"owner", "key_algo"},
			},
		},
	})
}

func TestAccGroupResource_WithKeyAlgo(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with RSA4096 key
			{
				Config: testAccGroupResourceConfig("testgrp2", "bastionadmin", "rsa4096"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_group.test",
						tfjsonpath.New("group"),
						knownvalue.StringExact("testgrp2"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group.test",
						tfjsonpath.New("key_algo"),
						knownvalue.StringExact("rsa4096"),
					),
				},
			},
		},
	})
}

func TestAccGroupResource_WithECDSA(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with ECDSA256 key
			{
				Config: testAccGroupResourceConfig("testgrp3", "bastionadmin", "ecdsa256"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_group.test",
						tfjsonpath.New("group"),
						knownvalue.StringExact("testgrp3"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group.test",
						tfjsonpath.New("key_algo"),
						knownvalue.StringExact("ecdsa256"),
					),
				},
			},
		},
	})
}

func TestAccGroupResource_RequiresReplace(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupResourceConfig("testgrp4", "bastionadmin", "ed25519"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_group.test",
						tfjsonpath.New("key_algo"),
						knownvalue.StringExact("ed25519"),
					),
				},
			},
			// Change key_algo should force replacement
			{
				Config: testAccGroupResourceConfig("testgrp4", "bastionadmin", "rsa2048"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bastion_group.test", "key_algo", "rsa2048"),
				),
			},
		},
	})
}

func TestAccGroupResource_ComputedAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and verify computed attributes
			{
				Config: testAccGroupResourceConfig("testgrp5", "bastionadmin", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bastion_group.test", "group", "testgrp5"),
					resource.TestCheckResourceAttr("bastion_group.test", "owner", "bastionadmin"),
					resource.TestCheckResourceAttr("bastion_group.test", "key_algo", "ed25519"),
					resource.TestCheckResourceAttrSet("bastion_group.test", "owners.#"),
					resource.TestCheckResourceAttrSet("bastion_group.test", "members.#"),
					resource.TestCheckResourceAttrSet("bastion_group.test", "gatekeepers.#"),
					resource.TestCheckResourceAttrSet("bastion_group.test", "aclkeepers.#"),
				),
			},
		},
	})
}

func TestAccGroupResource_WithModifyOptions(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with modify options
			{
				Config: testAccGroupResourceConfigWithModifyOptions("testgrp6", "bastionadmin", "", "totp", "900", "1800", "86400"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_group.test",
						tfjsonpath.New("group"),
						knownvalue.StringExact("testgrp6"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group.test",
						tfjsonpath.New("mfa_required"),
						knownvalue.StringExact("totp"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group.test",
						tfjsonpath.New("idle_lock_timeout"),
						knownvalue.StringExact("900"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group.test",
						tfjsonpath.New("idle_kill_timeout"),
						knownvalue.StringExact("1800"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group.test",
						tfjsonpath.New("guest_ttl_limit"),
						knownvalue.StringExact("86400"),
					),
				},
			},
			// Update modify options
			{
				Config: testAccGroupResourceConfigWithModifyOptions("testgrp6", "bastionadmin", "", "any", "1200", "2400", "43200"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_group.test",
						tfjsonpath.New("mfa_required"),
						knownvalue.StringExact("any"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group.test",
						tfjsonpath.New("idle_lock_timeout"),
						knownvalue.StringExact("1200"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group.test",
						tfjsonpath.New("idle_kill_timeout"),
						knownvalue.StringExact("2400"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group.test",
						tfjsonpath.New("guest_ttl_limit"),
						knownvalue.StringExact("43200"),
					),
				},
			},
			// ImportState testing with modify options
			{
				ResourceName:                         "bastion_group.test",
				ImportStateId:                        "testgrp6",
				ImportStateVerifyIdentifierAttribute: "group",
				ImportState:                          true,
				ImportStateVerify:                    true,
				// owner and key_algo are only used during creation
				ImportStateVerifyIgnore: []string{"owner", "key_algo"},
			},
		},
	})
}

func TestAccGroupResource_PartialModifyOptions(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create without modify options
			{
				Config: testAccGroupResourceConfig("testgrp7", "bastionadmin", ""),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_group.test",
						tfjsonpath.New("group"),
						knownvalue.StringExact("testgrp7"),
					),
				},
			},
			// Add only MFA requirement
			{
				Config: testAccGroupResourceConfigWithPartialOptions("testgrp7", "bastionadmin", "", map[string]any{
					"mfa_required": "password",
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_group.test",
						tfjsonpath.New("mfa_required"),
						knownvalue.StringExact("password"),
					),
				},
			},
			// Add idle timeouts
			{
				Config: testAccGroupResourceConfigWithPartialOptions("testgrp7", "bastionadmin", "", map[string]any{
					"mfa_required":      "password",
					"idle_lock_timeout": "600",
					"idle_kill_timeout": "1200",
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_group.test",
						tfjsonpath.New("mfa_required"),
						knownvalue.StringExact("password"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group.test",
						tfjsonpath.New("idle_lock_timeout"),
						knownvalue.StringExact("600"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group.test",
						tfjsonpath.New("idle_kill_timeout"),
						knownvalue.StringExact("1200"),
					),
				},
			},
		},
	})
}

func TestAccGroupResource_MFAPolicy(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test different MFA policies
			{
				Config: testAccGroupResourceConfigWithPartialOptions("testgrp8", "bastionadmin", "", map[string]any{
					"mfa_required": "none",
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_group.test",
						tfjsonpath.New("mfa_required"),
						knownvalue.StringExact("none"),
					),
				},
			},
			{
				Config: testAccGroupResourceConfigWithPartialOptions("testgrp8", "bastionadmin", "", map[string]any{
					"mfa_required": "totp",
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_group.test",
						tfjsonpath.New("mfa_required"),
						knownvalue.StringExact("totp"),
					),
				},
			},
			{
				Config: testAccGroupResourceConfigWithPartialOptions("testgrp8", "bastionadmin", "", map[string]any{
					"mfa_required": "password",
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_group.test",
						tfjsonpath.New("mfa_required"),
						knownvalue.StringExact("password"),
					),
				},
			},
			{
				Config: testAccGroupResourceConfigWithPartialOptions("testgrp8", "bastionadmin", "", map[string]any{
					"mfa_required": "any",
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_group.test",
						tfjsonpath.New("mfa_required"),
						knownvalue.StringExact("any"),
					),
				},
			},
		},
	})
}

// testAccGroupResourceConfig generates the Terraform configuration for testing.
func testAccGroupResourceConfig(groupName, owner, keyAlgo string) string { //nolint:unparam
	config := providerConfig()

	if keyAlgo != "" {
		config += fmt.Sprintf(`
resource "bastion_group" "test" {
  group    = %[1]q
  owner    = %[2]q
  key_algo = %[3]q
}
`, groupName, owner, keyAlgo)
	} else {
		config += fmt.Sprintf(`
resource "bastion_group" "test" {
  group = %[1]q
  owner = %[2]q
}
`, groupName, owner)
	}

	return config
}

// testAccGroupResourceConfigWithModifyOptions generates config with all modify options.
func testAccGroupResourceConfigWithModifyOptions(groupName, owner, keyAlgo, mfaRequired, idleLockTimeout, idleKillTimeout, guestTtlLimit string) string {
	config := providerConfig()

	keyAlgoStr := ""
	if keyAlgo != "" {
		keyAlgoStr = fmt.Sprintf("  key_algo = %q\n", keyAlgo)
	}

	config += fmt.Sprintf(`
resource "bastion_group" "test" {
  group             = %[1]q
  owner             = %[2]q
%[3]s mfa_required      = %[4]q
  idle_lock_timeout = %[5]q
  idle_kill_timeout = %[6]q
  guest_ttl_limit   = %[7]q
}
`, groupName, owner, keyAlgoStr, mfaRequired, idleLockTimeout, idleKillTimeout, guestTtlLimit)

	return config
}

// testAccGroupResourceConfigWithPartialOptions generates config with selected modify options.
func testAccGroupResourceConfigWithPartialOptions(groupName, owner, keyAlgo string, options map[string]any) string { //nolint:unparam
	config := providerConfig()

	resourceConfig := fmt.Sprintf(`
resource "bastion_group" "test" {
  group = %[1]q
  owner = %[2]q
`, groupName, owner)

	if keyAlgo != "" {
		resourceConfig += fmt.Sprintf("  key_algo = %q\n", keyAlgo)
	}

	if mfaRequired, ok := options["mfa_required"].(string); ok {
		resourceConfig += fmt.Sprintf("  mfa_required = %q\n", mfaRequired)
	}

	if idleLockTimeout, ok := options["idle_lock_timeout"].(string); ok {
		resourceConfig += fmt.Sprintf("  idle_lock_timeout = %q\n", idleLockTimeout)
	}

	if idleKillTimeout, ok := options["idle_kill_timeout"].(string); ok {
		resourceConfig += fmt.Sprintf("  idle_kill_timeout = %q\n", idleKillTimeout)
	}

	if guestTtlLimit, ok := options["guest_ttl_limit"].(string); ok {
		resourceConfig += fmt.Sprintf("  guest_ttl_limit = %q\n", guestTtlLimit)
	}

	resourceConfig += "}\n"
	config += resourceConfig

	return config
}
