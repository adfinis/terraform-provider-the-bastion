// Copyright (c) Adfinis
// SPDX-License-Identifier: GPL-3.0-or-later

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

func TestAccAccountResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountResourceConfig("testaccount1", true),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("account"),
						knownvalue.StringExact("testaccount1"),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:                         "bastion_account.test",
				ImportStateVerifyIdentifierAttribute: "account",
				ImportStateId:                        "testaccount1",
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
		},
	})
}

func TestAccAccountResource_WithSpecificUID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountResourceConfigWithUID("testaccount2", 9969),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("account"),
						knownvalue.StringExact("testaccount2"),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:                         "bastion_account.test",
				ImportStateVerifyIdentifierAttribute: "account",
				ImportStateId:                        "testaccount2",
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
		},
	})
}

func TestAccAccountResource_WithModifyOptions(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with modify options
			{
				Config: testAccAccountResourceConfigWithModifyOptions("testaccount3", true, 0, map[string]any{
					"always_active":     true,
					"osh_only":          false,
					"pam_auth_bypass":   true,
					"idle_ignore":       false,
					"max_inactive_days": 30,
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("account"),
						knownvalue.StringExact("testaccount3"),
					),
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("always_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("osh_only"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("pam_auth_bypass"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("idle_ignore"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("max_inactive_days"),
						knownvalue.Int64Exact(30),
					),
				},
			},
			// Update modify options
			{
				Config: testAccAccountResourceConfigWithModifyOptions("testaccount3", true, 0, map[string]any{
					"always_active":     false,
					"osh_only":          true,
					"pam_auth_bypass":   false,
					"idle_ignore":       true,
					"max_inactive_days": 60,
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("always_active"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("osh_only"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("pam_auth_bypass"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("idle_ignore"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("max_inactive_days"),
						knownvalue.Int64Exact(60),
					),
				},
			},
			// ImportState testing with modify options
			{
				ResourceName:                         "bastion_account.test",
				ImportStateId:                        "testaccount3",
				ImportStateVerifyIdentifierAttribute: "account",
				ImportState:                          true,
				ImportStateVerify:                    true,
				// uid and uid_auto are only used during creation
				ImportStateVerifyIgnore: []string{"uid", "uid_auto"},
			},
		},
	})
}

func TestAccAccountResource_MFASettings(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create without MFA settings
			{
				Config: testAccAccountResourceConfig("testaccount4", true),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("account"),
						knownvalue.StringExact("testaccount4"),
					),
				},
			},
			// Add MFA settings
			{
				Config: testAccAccountResourceConfigWithModifyOptions("testaccount4", true, 0, map[string]any{
					"mfa_password_required": "yes",
					"mfa_totp_required":     "no",
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("mfa_password_required"),
						knownvalue.StringExact("yes"),
					),
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("mfa_totp_required"),
						knownvalue.StringExact("no"),
					),
				},
			},
			// Update to bypass
			{
				Config: testAccAccountResourceConfigWithModifyOptions("testaccount4", true, 0, map[string]any{
					"mfa_password_required": "bypass",
					"mfa_totp_required":     "bypass",
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("mfa_password_required"),
						knownvalue.StringExact("bypass"),
					),
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("mfa_totp_required"),
						knownvalue.StringExact("bypass"),
					),
				},
			},
		},
	})
}

func TestAccAccountResource_PersonalEgressMFA(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test different personal egress MFA policies
			{
				Config: testAccAccountResourceConfigWithModifyOptions("testaccount5", true, 0, map[string]any{
					"personal_egress_mfa_required": "none",
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("personal_egress_mfa_required"),
						knownvalue.StringExact("none"),
					),
				},
			},
			{
				Config: testAccAccountResourceConfigWithModifyOptions("testaccount5", true, 0, map[string]any{
					"personal_egress_mfa_required": "password",
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("personal_egress_mfa_required"),
						knownvalue.StringExact("password"),
					),
				},
			},
			{
				Config: testAccAccountResourceConfigWithModifyOptions("testaccount5", true, 0, map[string]any{
					"personal_egress_mfa_required": "totp",
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("personal_egress_mfa_required"),
						knownvalue.StringExact("totp"),
					),
				},
			},
			{
				Config: testAccAccountResourceConfigWithModifyOptions("testaccount5", true, 0, map[string]any{
					"personal_egress_mfa_required": "any",
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("personal_egress_mfa_required"),
						knownvalue.StringExact("any"),
					),
				},
			},
		},
	})
}

func TestAccAccountResource_EgressSettings(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with egress settings
			{
				Config: testAccAccountResourceConfigWithModifyOptions("testaccount6", true, 0, map[string]any{
					"egress_strict_host_key_checking": "yes",
					"egress_session_multiplexing":     "yes",
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("egress_strict_host_key_checking"),
						knownvalue.StringExact("yes"),
					),
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("egress_session_multiplexing"),
						knownvalue.StringExact("yes"),
					),
				},
			},
			// Update egress settings
			{
				Config: testAccAccountResourceConfigWithModifyOptions("testaccount6", true, 0, map[string]any{
					"egress_strict_host_key_checking": "accept-new",
					"egress_session_multiplexing":     "no",
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("egress_strict_host_key_checking"),
						knownvalue.StringExact("accept-new"),
					),
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("egress_session_multiplexing"),
						knownvalue.StringExact("no"),
					),
				},
			},
			// Test default and bypass values
			{
				Config: testAccAccountResourceConfigWithModifyOptions("testaccount6", true, 0, map[string]any{
					"egress_strict_host_key_checking": "default",
					"egress_session_multiplexing":     "default",
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("egress_strict_host_key_checking"),
						knownvalue.StringExact("default"),
					),
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("egress_session_multiplexing"),
						knownvalue.StringExact("default"),
					),
				},
			},
		},
	})
}

func TestAccAccountResource_PartialModifyOptions(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create without modify options
			{
				Config: testAccAccountResourceConfig("testaccount7", true),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("account"),
						knownvalue.StringExact("testaccount7"),
					),
				},
			},
			// Add only always_active
			{
				Config: testAccAccountResourceConfigWithModifyOptions("testaccount7", true, 0, map[string]any{
					"always_active": true,
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("always_active"),
						knownvalue.Bool(true),
					),
				},
			},
			// Add more options gradually
			{
				Config: testAccAccountResourceConfigWithModifyOptions("testaccount7", true, 0, map[string]any{
					"always_active":   true,
					"pam_auth_bypass": true,
					"idle_ignore":     true,
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("always_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("pam_auth_bypass"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("idle_ignore"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func TestAccAccountResource_PubkeyAuthOptional(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with pubkey_auth_optional set to false
			{
				Config: testAccAccountResourceConfigWithModifyOptions("testaccount8", true, 0, map[string]any{
					"pubkey_auth_optional": false,
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("pubkey_auth_optional"),
						knownvalue.Bool(false),
					),
				},
			},
			// Update to true
			{
				Config: testAccAccountResourceConfigWithModifyOptions("testaccount8", true, 0, map[string]any{
					"pubkey_auth_optional": true,
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("pubkey_auth_optional"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func TestAccAccountResource_ComplexUpdate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with comprehensive options
			{
				Config: testAccAccountResourceConfigWithModifyOptions("testaccount9", false, 9942, map[string]any{
					"always_active":                   true,
					"osh_only":                        false,
					"max_inactive_days":               30,
					"pam_auth_bypass":                 true,
					"mfa_password_required":           "no",
					"mfa_totp_required":               "yes",
					"egress_strict_host_key_checking": "accept-new",
					"egress_session_multiplexing":     "yes",
					"personal_egress_mfa_required":    "totp",
					"idle_ignore":                     false,
					"pubkey_auth_optional":            false,
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("always_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("osh_only"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("max_inactive_days"),
						knownvalue.Int64Exact(30),
					),
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("mfa_totp_required"),
						knownvalue.StringExact("yes"),
					),
				},
			},
			// Update multiple options
			{
				Config: testAccAccountResourceConfigWithModifyOptions("testaccount9", false, 9942, map[string]any{
					"always_active":                   false,
					"osh_only":                        true,
					"max_inactive_days":               60,
					"pam_auth_bypass":                 false,
					"mfa_password_required":           "bypass",
					"mfa_totp_required":               "bypass",
					"egress_strict_host_key_checking": "yes",
					"egress_session_multiplexing":     "no",
					"personal_egress_mfa_required":    "any",
					"idle_ignore":                     true,
					"pubkey_auth_optional":            true,
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("always_active"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("osh_only"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("max_inactive_days"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("mfa_password_required"),
						knownvalue.StringExact("bypass"),
					),
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("mfa_totp_required"),
						knownvalue.StringExact("bypass"),
					),
					statecheck.ExpectKnownValue(
						"bastion_account.test",
						tfjsonpath.New("personal_egress_mfa_required"),
						knownvalue.StringExact("any"),
					),
				},
			},
		},
	})
}

// testAccAccountResourceConfig generates the Terraform configuration for testing with uid_auto.
func testAccAccountResourceConfig(accountName string, uidAuto bool) string {
	config := providerConfig

	if uidAuto {
		config += fmt.Sprintf(`
resource "bastion_account" "test" {
  account  = %[1]q
  uid_auto = true
  public_key = %[2]q
}
`, accountName, string(testutils.SSHPublicKey))
	} else {
		config += fmt.Sprintf(`
resource "bastion_account" "test" {
  account = %[1]q
  no_key  = true
}
`, accountName)
	}

	return config
}

// testAccAccountResourceConfigWithUID generates config with a specific UID.
func testAccAccountResourceConfigWithUID(accountName string, uid int) string {
	config := providerConfig

	config += fmt.Sprintf(`
resource "bastion_account" "test" {
  account = %[1]q
  uid     = %[2]d
  no_key  = true
}
`, accountName, uid)

	return config
}

// testAccAccountResourceConfigWithModifyOptions generates config with selected modify options.
func testAccAccountResourceConfigWithModifyOptions(accountName string, uidAuto bool, uid int, options map[string]any) string {
	config := providerConfig

	resourceConfig := fmt.Sprintf(`
resource "bastion_account" "test" {
  account = %[1]q
  no_key  = true
`, accountName)

	if uidAuto {
		resourceConfig += "  uid_auto = true\n"
	}
	if !uidAuto && uid != 0 {
		resourceConfig += fmt.Sprintf("  uid     = %d\n", uid)
	}

	if alwaysActive, ok := options["always_active"].(bool); ok {
		resourceConfig += fmt.Sprintf("  always_active = %t\n", alwaysActive)
	}

	if oshOnly, ok := options["osh_only"].(bool); ok {
		resourceConfig += fmt.Sprintf("  osh_only = %t\n", oshOnly)
	}

	if maxInactiveDays, ok := options["max_inactive_days"].(int); ok {
		resourceConfig += fmt.Sprintf("  max_inactive_days = %d\n", maxInactiveDays)
	}

	if pamAuthBypass, ok := options["pam_auth_bypass"].(bool); ok {
		resourceConfig += fmt.Sprintf("  pam_auth_bypass = %t\n", pamAuthBypass)
	}

	if mfaPasswordRequired, ok := options["mfa_password_required"].(string); ok {
		resourceConfig += fmt.Sprintf("  mfa_password_required = %q\n", mfaPasswordRequired)
	}

	if mfaTOTPRequired, ok := options["mfa_totp_required"].(string); ok {
		resourceConfig += fmt.Sprintf("  mfa_totp_required = %q\n", mfaTOTPRequired)
	}

	if egressStrictHostKeyChecking, ok := options["egress_strict_host_key_checking"].(string); ok {
		resourceConfig += fmt.Sprintf("  egress_strict_host_key_checking = %q\n", egressStrictHostKeyChecking)
	}

	if egressSessionMultiplexing, ok := options["egress_session_multiplexing"].(string); ok {
		resourceConfig += fmt.Sprintf("  egress_session_multiplexing = %q\n", egressSessionMultiplexing)
	}

	if personalEgressMFARequired, ok := options["personal_egress_mfa_required"].(string); ok {
		resourceConfig += fmt.Sprintf("  personal_egress_mfa_required = %q\n", personalEgressMFARequired)
	}

	if idleIgnore, ok := options["idle_ignore"].(bool); ok {
		resourceConfig += fmt.Sprintf("  idle_ignore = %t\n", idleIgnore)
	}

	if pubkeyAuthOptional, ok := options["pubkey_auth_optional"].(bool); ok {
		resourceConfig += fmt.Sprintf("  pubkey_auth_optional = %t\n", pubkeyAuthOptional)
	}

	resourceConfig += "}\n"
	config += resourceConfig

	return config
}
