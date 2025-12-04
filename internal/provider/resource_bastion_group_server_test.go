// Copyright (c) Adfinis
// SPDX-License-Identifier: GPL-3.0-or-later

package provider

import (
	"fmt"
	"testing"

	"github.com/adfinis/terraform-provider-bastion/bastion"
	"github.com/adfinis/terraform-provider-bastion/internal/provider/testutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/assert"
)

func TestAccGroupServerResource(t *testing.T) {
	err := testutils.CreateGroup("testgrpsrv1", "bastionadmin", bastion.ED25519)
	if err != nil {
		t.Errorf("Unable to create test group: %s", err)
	}

	t.Cleanup(func() {
		err := testutils.DeleteGroup("testgrpsrv1")
		if err != nil {
			t.Errorf("Unable to delete test group: %s", err)
		}
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccGroupServerResourceConfig("testgrpsrv1", "192.168.1.100", "22", "root", "", "", "", ""),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("group"),
						knownvalue.StringExact("testgrpsrv1"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("ip"),
						knownvalue.StringExact("192.168.1.100"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("port"),
						knownvalue.StringExact("22"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("user"),
						knownvalue.StringExact("root"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("testgrpsrv1:192.168.1.100:22:root"),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "bastion_group_server.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "testgrpsrv1:192.168.1.100:22:root",
				// force is only used during creation
				ImportStateVerifyIgnore: []string{"force"},
			},
		},
	})
}

func TestAccGroupServerResource_Wildcard(t *testing.T) {
	err := testutils.CreateGroup("testgrpsrv2", "bastionadmin", bastion.ED25519)
	if err != nil {
		t.Errorf("Unable to create test group: %s", err)
	}

	t.Cleanup(func() {
		err := testutils.DeleteGroup("testgrpsrv2")
		if err != nil {
			t.Errorf("Unable to delete test group: %s", err)
		}
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with wildcard port and user
			{
				Config: testAccGroupServerResourceConfig("testgrpsrv2", "192.168.1.101", "*", "*", "", "", "", ""),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("ip"),
						knownvalue.StringExact("192.168.1.101"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("port"),
						knownvalue.StringExact("*"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("user"),
						knownvalue.StringExact("*"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("testgrpsrv2:192.168.1.101:*:*"),
					),
				},
			},
		},
	})
}

func TestAccGroupServerResource_WithComment(t *testing.T) {
	err := testutils.CreateGroup("testgrpsrv3", "bastionadmin", bastion.ED25519)
	if err != nil {
		t.Errorf("Unable to create test group: %s", err)
	}

	t.Cleanup(func() {
		err := testutils.DeleteGroup("testgrpsrv3")
		if err != nil {
			t.Errorf("Unable to delete test group: %s", err)
		}
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with comment
			{
				Config: testAccGroupServerResourceConfig("testgrpsrv3", "192.168.1.102", "22", "ubuntu", "Test server access", "", "", ""),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("ip"),
						knownvalue.StringExact("192.168.1.102"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("comment"),
						knownvalue.StringExact("Test server access"),
					),
				},
			},
		},
	})
}

func TestAccGroupServerResource_WithProxy(t *testing.T) {
	err := testutils.CreateGroup("testgrpsrv4", "bastionadmin", bastion.ED25519)
	if err != nil {
		t.Errorf("Unable to create test group: %s", err)
	}

	t.Cleanup(func() {
		err := testutils.DeleteGroup("testgrpsrv4")
		if err != nil {
			t.Errorf("Unable to delete test group: %s", err)
		}
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with proxy settings
			{
				Config: testAccGroupServerResourceConfig("testgrpsrv4", "10.0.0.50", "22", "admin", "", "192.168.1.1", "22", "proxy_user"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("ip"),
						knownvalue.StringExact("10.0.0.50"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("proxy_ip"),
						knownvalue.StringExact("192.168.1.1"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("proxy_port"),
						knownvalue.StringExact("22"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("proxy_user"),
						knownvalue.StringExact("proxy_user"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("testgrpsrv4:10.0.0.50:22:admin:192.168.1.1:22:proxy_user"),
					),
				},
			},
			// ImportState testing with proxy
			{
				ResourceName:            "bastion_group_server.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           "testgrpsrv4:10.0.0.50:22:admin:192.168.1.1:22:proxy_user",
				ImportStateVerifyIgnore: []string{"force"},
			},
		},
	})
}

func TestAccGroupServerResource_Multiple(t *testing.T) {
	err := testutils.CreateGroup("testgrpsrv6", "bastionadmin", bastion.ED25519)
	if err != nil {
		t.Errorf("Unable to create test group: %s", err)
	}

	t.Cleanup(func() {
		err := testutils.DeleteGroup("testgrpsrv6")
		if err != nil {
			t.Errorf("Unable to delete test group: %s", err)
		}
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create multiple server accesses
			{
				Config: testAccGroupServerResourceConfigMultiple("testgrpsrv6"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bastion_group_server.test1", "group", "testgrpsrv6"),
					resource.TestCheckResourceAttr("bastion_group_server.test1", "ip", "192.168.1.200"),
					resource.TestCheckResourceAttr("bastion_group_server.test1", "port", "22"),
					resource.TestCheckResourceAttr("bastion_group_server.test1", "user", "root"),
					resource.TestCheckResourceAttr("bastion_group_server.test2", "group", "testgrpsrv6"),
					resource.TestCheckResourceAttr("bastion_group_server.test2", "ip", "192.168.1.201"),
					resource.TestCheckResourceAttr("bastion_group_server.test2", "port", "2222"),
					resource.TestCheckResourceAttr("bastion_group_server.test2", "user", "admin"),
				),
			},
		},
	})
}

func TestAccGroupServerResource_RequiresReplace(t *testing.T) {
	err := testutils.CreateGroup("testgrpsrv7", "bastionadmin", bastion.ED25519)
	if err != nil {
		t.Errorf("Unable to create test group: %s", err)
	}

	t.Cleanup(func() {
		err := testutils.DeleteGroup("testgrpsrv7")
		if err != nil {
			t.Errorf("Unable to delete test group: %s", err)
		}
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupServerResourceConfig("testgrpsrv7", "192.168.1.210", "22", "user1", "", "", "", ""),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("ip"),
						knownvalue.StringExact("192.168.1.210"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("user"),
						knownvalue.StringExact("user1"),
					),
				},
			},
			// Change ip should force replacement
			{
				Config: testAccGroupServerResourceConfig("testgrpsrv7", "192.168.1.211", "22", "user1", "", "", "", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bastion_group_server.test", "ip", "192.168.1.211"),
					resource.TestCheckResourceAttr("bastion_group_server.test", "id", "testgrpsrv7:192.168.1.211:22:user1"),
				),
			},
		},
	})
}

func TestAccGroupServerResource_Subnet(t *testing.T) {
	err := testutils.CreateGroup("testgrpsrv9", "bastionadmin", bastion.ED25519)
	if err != nil {
		t.Errorf("Unable to create test group: %s", err)
	}

	t.Cleanup(func() {
		err := testutils.DeleteGroup("testgrpsrv9")
		if err != nil {
			t.Errorf("Unable to delete test group: %s", err)
		}
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with subnet
			{
				Config: testAccGroupServerResourceConfig("testgrpsrv9", "192.168.2.0/24", "22", "root", "", "", "", ""),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("ip"),
						knownvalue.StringExact("192.168.2.0/24"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("port"),
						knownvalue.StringExact("22"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("user"),
						knownvalue.StringExact("root"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("testgrpsrv9:192.168.2.0/24:22:root"),
					),
				},
			},
		},
	})
}

func TestAccGroupServerResource_WithTTL(t *testing.T) {
	err := testutils.CreateGroup("testgrpsrv10", "bastionadmin", bastion.ED25519)
	if err != nil {
		t.Errorf("Unable to create test group: %s", err)
	}

	t.Cleanup(func() {
		err := testutils.DeleteGroup("testgrpsrv10")
		if err != nil {
			t.Errorf("Unable to delete test group: %s", err)
		}
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with TTL
			{
				Config: testAccGroupServerResourceConfigWithTTL("testgrpsrv10", "192.168.1.150", "22", "root", 3600),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("group"),
						knownvalue.StringExact("testgrpsrv10"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("ip"),
						knownvalue.StringExact("192.168.1.150"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("port"),
						knownvalue.StringExact("22"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("user"),
						knownvalue.StringExact("root"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("ttl"),
						knownvalue.Int64Exact(3600),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("testgrpsrv10:192.168.1.150:22:root"),
					),
				},
			},
		},
	})
}

func TestAccGroupServerResource_WithForceKey(t *testing.T) {
	groupName := "testgrpsrv11"
	err := testutils.CreateGroup(groupName, "bastionadmin", bastion.ED25519)
	if err != nil {
		t.Errorf("Unable to create test group: %s", err)
	}

	t.Cleanup(func() {
		err := testutils.DeleteGroup(groupName)
		if err != nil {
			t.Errorf("Unable to delete test group: %s", err)
		}
	})

	fingerprint, err := testutils.GetGroupKeyFingerprint(groupName)
	if err != nil {
		t.Errorf("Unable to get group key fingerprint: %s", err)
	}
	if fingerprint == "" {
		t.Skip("No SSH key fingerprint found for group")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with force_key
			{
				Config: testAccGroupServerResourceConfigWithForceKey(groupName, "192.168.1.160", "22", "admin", fingerprint),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("group"),
						knownvalue.StringExact(groupName),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("ip"),
						knownvalue.StringExact("192.168.1.160"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("port"),
						knownvalue.StringExact("22"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("user"),
						knownvalue.StringExact("admin"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("force_key"),
						knownvalue.StringExact(fingerprint),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact(groupName+":192.168.1.160:22:admin"),
					),
				},
			},
		},
	})
}

func TestParseImportID(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "basic IPv4 without proxy",
			input:    "mygroup:192.168.1.100:22:root",
			expected: []string{"mygroup", "192.168.1.100", "22", "root"},
		},
		{
			name:     "IPv4 with proxy",
			input:    "mygroup:192.168.1.100:22:root:10.0.0.1:22:proxy_user",
			expected: []string{"mygroup", "192.168.1.100", "22", "root", "10.0.0.1", "22", "proxy_user"},
		},
		{
			name:     "IPv6 loopback without proxy",
			input:    "mygroup:[::1]:22:root",
			expected: []string{"mygroup", "::1", "22", "root"},
		},
		{
			name:     "IPv6 full address without proxy",
			input:    "mygroup:[2001:db8::1]:22:root",
			expected: []string{"mygroup", "2001:db8::1", "22", "root"},
		},
		{
			name:     "IPv6 with proxy IPv4",
			input:    "mygroup:[2001:db8::1]:22:root:192.168.1.1:22:proxy_user",
			expected: []string{"mygroup", "2001:db8::1", "22", "root", "192.168.1.1", "22", "proxy_user"},
		},
		{
			name:     "IPv6 with proxy IPv6",
			input:    "mygroup:[2001:db8::1]:22:root:[fd00::1]:22:proxy_user",
			expected: []string{"mygroup", "2001:db8::1", "22", "root", "fd00::1", "22", "proxy_user"},
		},
		{
			name:     "IPv6 with wildcard port and user",
			input:    "mygroup:[fe80::1]:*:*",
			expected: []string{"mygroup", "fe80::1", "*", "*"},
		},
		{
			name:     "IPv4 subnet without proxy",
			input:    "mygroup:192.168.0.0/24:22:root",
			expected: []string{"mygroup", "192.168.0.0/24", "22", "root"},
		},
		{
			name:     "IPv6 subnet without proxy",
			input:    "mygroup:[2001:db8::/32]:22:root",
			expected: []string{"mygroup", "2001:db8::/32", "22", "root"},
		},
		{
			name:     "complex IPv6 address",
			input:    "mygroup:[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:2222:admin",
			expected: []string{"mygroup", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", "2222", "admin"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := parseImportID(tc.input)
			assert.Equal(t, tc.expected, result, "parseImportID(%q) should return %v, got %v", tc.input, tc.expected, result)
		})
	}
}

func TestAccGroupServerResource_IPv6(t *testing.T) {
	err := testutils.CreateGroup("testgrpsrvipv6", "bastionadmin", bastion.ED25519)
	if err != nil {
		t.Errorf("Unable to create test group: %s", err)
	}

	t.Cleanup(func() {
		err := testutils.DeleteGroup("testgrpsrvipv6")
		if err != nil {
			t.Errorf("Unable to delete test group: %s", err)
		}
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with IPv6 address
			{
				Config: testAccGroupServerResourceConfig("testgrpsrvipv6", "::1", "22", "root", "", "", "", ""),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("group"),
						knownvalue.StringExact("testgrpsrvipv6"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("ip"),
						knownvalue.StringExact("::1"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("port"),
						knownvalue.StringExact("22"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("user"),
						knownvalue.StringExact("root"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("testgrpsrvipv6:[::1]:22:root"),
					),
				},
			},
			// ImportState testing with IPv6
			{
				ResourceName:            "bastion_group_server.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           "testgrpsrvipv6:[::1]:22:root",
				ImportStateVerifyIgnore: []string{"force"},
			},
		},
	})
}

func TestAccGroupServerResource_IPv6WithProxy(t *testing.T) {
	err := testutils.CreateGroup("testgrpsrvipv6prxy", "bastionadmin", bastion.ED25519)
	if err != nil {
		t.Errorf("Unable to create test group: %s", err)
	}

	t.Cleanup(func() {
		err := testutils.DeleteGroup("testgrpsrvipv6prxy")
		if err != nil {
			t.Errorf("Unable to delete test group: %s", err)
		}
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with IPv6 target and IPv6 proxy
			{
				Config: testAccGroupServerResourceConfig("testgrpsrvipv6prxy", "2001:db8::1", "22", "admin", "", "fd00::1", "22", "proxy_user"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("ip"),
						knownvalue.StringExact("2001:db8::1"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("proxy_ip"),
						knownvalue.StringExact("fd00::1"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("proxy_port"),
						knownvalue.StringExact("22"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("proxy_user"),
						knownvalue.StringExact("proxy_user"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("testgrpsrvipv6prxy:[2001:db8::1]:22:admin:[fd00::1]:22:proxy_user"),
					),
				},
			},
			// ImportState testing with IPv6 and proxy
			{
				ResourceName:            "bastion_group_server.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           "testgrpsrvipv6prxy:[2001:db8::1]:22:admin:[fd00::1]:22:proxy_user",
				ImportStateVerifyIgnore: []string{"force"},
			},
		},
	})
}

func TestAccGroupServerResource_IPv6Subnet(t *testing.T) {
	err := testutils.CreateGroup("testgrpsrvipv6sbn", "bastionadmin", bastion.ED25519)
	if err != nil {
		t.Errorf("Unable to create test group: %s", err)
	}

	t.Cleanup(func() {
		err := testutils.DeleteGroup("testgrpsrvipv6sbn")
		if err != nil {
			t.Errorf("Unable to delete test group: %s", err)
		}
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with IPv6 subnet
			{
				Config: testAccGroupServerResourceConfig("testgrpsrvipv6sbn", "2001:db8::/32", "22", "root", "", "", "", ""),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("ip"),
						knownvalue.StringExact("2001:db8::/32"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_server.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("testgrpsrvipv6sbn:[2001:db8::/32]:22:root"),
					),
				},
			},
		},
	})
}

// testAccGroupServerResourceConfig generates the Terraform configuration for testing.
func testAccGroupServerResourceConfig(groupName, ip, port, user, comment, proxyIP, proxyPort, proxyUser string) string {
	config := providerConfig
	config += fmt.Sprintf(`
resource "bastion_group_server" "test" {
  group = %[1]q
  ip  = %[2]q
  port  = %[3]q
  user  = %[4]q
  force = true
`, groupName, ip, port, user)

	if comment != "" {
		config += fmt.Sprintf("  comment = %q\n", comment)
	}

	if proxyIP != "" {
		config += fmt.Sprintf("  proxy_ip   = %q\n", proxyIP)
		config += fmt.Sprintf("  proxy_port = %q\n", proxyPort)
		config += fmt.Sprintf("  proxy_user = %q\n", proxyUser)
	}

	config += "}\n"

	return config
}

// testAccGroupServerResourceConfigMultiple generates config with multiple server accesses.
func testAccGroupServerResourceConfigMultiple(groupName string) string {
	config := providerConfig
	config += fmt.Sprintf(`
resource "bastion_group_server" "test1" {
  group = %[1]q
  ip  = "192.168.1.200"
  port  = "22"
  user  = "root"
  force = true
}

resource "bastion_group_server" "test2" {
  group = %[1]q
  ip  = "192.168.1.201"
  port  = "2222"
  user  = "admin"
  force = true
}
`, groupName)

	return config
}

// testAccGroupServerResourceConfigWithTTL generates config with TTL.
func testAccGroupServerResourceConfigWithTTL(groupName, ip, port, user string, ttl int64) string {
	config := providerConfig
	config += fmt.Sprintf(`
resource "bastion_group_server" "test" {
  group = %[1]q
  ip    = %[2]q
  port  = %[3]q
  user  = %[4]q
  ttl   = %[5]d
  force = true
}
`, groupName, ip, port, user, ttl)

	return config
}

// testAccGroupServerResourceConfigWithForceKey generates config with force_key.
func testAccGroupServerResourceConfigWithForceKey(groupName, ip, port, user, forceKey string) string {
	config := providerConfig
	config += fmt.Sprintf(`
resource "bastion_group_server" "test" {
  group     = %[1]q
  ip        = %[2]q
  port      = %[3]q
  user      = %[4]q
  force_key = %[5]q
  force     = true
}
`, groupName, ip, port, user, forceKey)

	return config
}
