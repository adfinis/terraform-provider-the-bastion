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
)

func TestAccGroupGuestAccessResource(t *testing.T) {
	groupName := "testgrpguest1"
	accountName := "testguestaccount1"

	err := testutils.CreateGroup(groupName, "bastionadmin", bastion.ED25519)
	if err != nil {
		t.Errorf("Unable to create test group: %s", err)
	}
	err = testutils.CreateAccount(accountName)
	if err != nil {
		t.Errorf("Unable to create test account: %s", err)
	}
	err = testutils.CreateGroupServerAccess(groupName, "192.168.1.100", "22", "root")
	if err != nil {
		t.Errorf("Unable to create test server access: %s", err)
	}

	t.Cleanup(func() {
		err := testutils.DeleteGroupServerAccess(groupName, "192.168.1.100", "22", "root")
		if err != nil {
			t.Errorf("Unable to delete test server access: %s", err)
		}
		err = testutils.DeleteAccount(accountName)
		if err != nil {
			t.Errorf("Unable to delete test account: %s", err)
		}
		err = testutils.DeleteGroup(groupName)
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
				Config: testAccGroupGuestAccessResourceConfig(groupName, accountName, "192.168.1.100", "22", "root", "", ""),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_group_guest_access.test",
						tfjsonpath.New("group"),
						knownvalue.StringExact(groupName),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_guest_access.test",
						tfjsonpath.New("account"),
						knownvalue.StringExact(accountName),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_guest_access.test",
						tfjsonpath.New("ip"),
						knownvalue.StringExact("192.168.1.100"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_guest_access.test",
						tfjsonpath.New("port"),
						knownvalue.StringExact("22"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_guest_access.test",
						tfjsonpath.New("user"),
						knownvalue.StringExact("root"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_guest_access.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact(fmt.Sprintf("%s:%s:192.168.1.100:22:root", groupName, accountName)),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "bastion_group_guest_access.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     fmt.Sprintf("%s:%s:192.168.1.100:22:root", groupName, accountName),
			},
		},
	})
}

func TestAccGroupGuestAccessResource_WithComment(t *testing.T) {
	groupName := "testgrpguest2"
	accountName := "testguestaccount2"

	err := testutils.CreateGroup(groupName, "bastionadmin", bastion.ED25519)
	if err != nil {
		t.Errorf("Unable to create test group: %s", err)
	}
	err = testutils.CreateAccount(accountName)
	if err != nil {
		t.Errorf("Unable to create test account: %s", err)
	}
	err = testutils.CreateGroupServerAccess(groupName, "192.168.1.101", "22", "ubuntu")
	if err != nil {
		t.Errorf("Unable to create test server access: %s", err)
	}

	t.Cleanup(func() {
		err := testutils.DeleteGroupServerAccess(groupName, "192.168.1.101", "22", "ubuntu")
		if err != nil {
			t.Errorf("Unable to delete test server access: %s", err)
		}
		err = testutils.DeleteAccount(accountName)
		if err != nil {
			t.Errorf("Unable to delete test account: %s", err)
		}
		err = testutils.DeleteGroup(groupName)
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
				Config: testAccGroupGuestAccessResourceConfig(groupName, accountName, "192.168.1.101", "22", "ubuntu", "Test guest access", ""),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_group_guest_access.test",
						tfjsonpath.New("comment"),
						knownvalue.StringExact("Test guest access"),
					),
				},
			},
		},
	})
}

func TestAccGroupGuestAccessResource_Protocol(t *testing.T) {
	groupName := "testgrpguest3"
	accountName := "testguestaccount3"

	err := testutils.CreateGroup(groupName, "bastionadmin", bastion.ED25519)
	if err != nil {
		t.Errorf("Unable to create test group: %s", err)
	}
	err = testutils.CreateAccount(accountName)
	if err != nil {
		t.Errorf("Unable to create test account: %s", err)
	}
	err = testutils.CreateGroupServerAccessWithProtocol(groupName, "192.168.1.102", "22", "sftp")
	if err != nil {
		t.Errorf("Unable to create test server access: %s", err)
	}

	t.Cleanup(func() {
		err := testutils.DeleteGroupServerAccessWithProtocol(groupName, "192.168.1.102", "22", "sftp")
		if err != nil {
			t.Errorf("Unable to delete test server access: %s", err)
		}
		err = testutils.DeleteAccount(accountName)
		if err != nil {
			t.Errorf("Unable to delete test account: %s", err)
		}
		err = testutils.DeleteGroup(groupName)
		if err != nil {
			t.Errorf("Unable to delete test group: %s", err)
		}
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with protocol
			{
				Config: testAccGroupGuestAccessResourceConfigWithProtocol(groupName, accountName, "192.168.1.102", "22", "sftp", ""),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_group_guest_access.test",
						tfjsonpath.New("protocol"),
						knownvalue.StringExact("sftp"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_guest_access.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact(fmt.Sprintf("%s:%s:192.168.1.102:22::sftp", groupName, accountName)),
					),
				},
			},
		},
	})
}

func TestAccGroupGuestAccessResource_RequiresReplace(t *testing.T) {
	groupName := "testgrpguest4"
	accountName := "testguestaccount4"
	account2Name := "testguestaccount4b"

	err := testutils.CreateGroup(groupName, "bastionadmin", bastion.ED25519)
	if err != nil {
		t.Errorf("Unable to create test group: %s", err)
	}
	err = testutils.CreateAccounts(accountName, account2Name)
	if err != nil {
		t.Errorf("Unable to create test accounts: %s", err)
	}
	err = testutils.CreateGroupServerAccess(groupName, "192.168.1.103", "22", "admin")
	if err != nil {
		t.Errorf("Unable to create test server access: %s", err)
	}

	t.Cleanup(func() {
		err := testutils.DeleteGroupServerAccess(groupName, "192.168.1.103", "22", "admin")
		if err != nil {
			t.Errorf("Unable to delete test server access: %s", err)
		}
		err = testutils.DeleteAccounts(accountName, account2Name)
		if err != nil {
			t.Errorf("Unable to delete test accounts: %s", err)
		}
		err = testutils.DeleteGroup(groupName)
		if err != nil {
			t.Errorf("Unable to delete test group: %s", err)
		}
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccGroupGuestAccessResourceConfig(groupName, accountName, "192.168.1.103", "22", "admin", "", ""),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_group_guest_access.test",
						tfjsonpath.New("account"),
						knownvalue.StringExact(accountName),
					),
				},
			},
			// Change account (requires replace)
			{
				Config: testAccGroupGuestAccessResourceConfig(groupName, account2Name, "192.168.1.103", "22", "admin", "", ""),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_group_guest_access.test",
						tfjsonpath.New("account"),
						knownvalue.StringExact(account2Name),
					),
				},
			},
		},
	})
}

func TestAccGroupGuestAccessResource_IPv6(t *testing.T) {
	groupName := "testgrpguest5"
	accountName := "testguestaccount5"

	err := testutils.CreateGroup(groupName, "bastionadmin", bastion.ED25519)
	if err != nil {
		t.Errorf("Unable to create test group: %s", err)
	}
	err = testutils.CreateAccount(accountName)
	if err != nil {
		t.Errorf("Unable to create test account: %s", err)
	}
	err = testutils.CreateGroupServerAccess(groupName, "::1", "22", "root")
	if err != nil {
		t.Errorf("Unable to create test server access: %s", err)
	}

	t.Cleanup(func() {
		err := testutils.DeleteGroupServerAccess(groupName, "::1", "22", "root")
		if err != nil {
			t.Errorf("Unable to delete test server access: %s", err)
		}
		err = testutils.DeleteAccount(accountName)
		if err != nil {
			t.Errorf("Unable to delete test account: %s", err)
		}
		err = testutils.DeleteGroup(groupName)
		if err != nil {
			t.Errorf("Unable to delete test group: %s", err)
		}
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with IPv6
			{
				Config: testAccGroupGuestAccessResourceConfig(groupName, accountName, "::1", "22", "root", "", ""),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_group_guest_access.test",
						tfjsonpath.New("ip"),
						knownvalue.StringExact("::1"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_guest_access.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact(fmt.Sprintf("%s:%s:[::1]:22:root", groupName, accountName)),
					),
				},
			},
		},
	})
}

func TestAccGroupGuestAccessResource_Multiple(t *testing.T) {
	groupName := "testgrpguest6"
	accountName := "testguestaccount6"

	err := testutils.CreateGroup(groupName, "bastionadmin", bastion.ED25519)
	if err != nil {
		t.Errorf("Unable to create test group: %s", err)
	}
	err = testutils.CreateAccount(accountName)
	if err != nil {
		t.Errorf("Unable to create test account: %s", err)
	}
	err = testutils.CreateGroupServerAccess(groupName, "192.168.1.104", "22", "root")
	if err != nil {
		t.Errorf("Unable to create test server access 1: %s", err)
	}
	err = testutils.CreateGroupServerAccess(groupName, "192.168.1.105", "22", "admin")
	if err != nil {
		t.Errorf("Unable to create test server access 2: %s", err)
	}

	t.Cleanup(func() {
		err := testutils.DeleteGroupServerAccess(groupName, "192.168.1.104", "22", "root")
		if err != nil {
			t.Errorf("Unable to delete test server access 1: %s", err)
		}
		err = testutils.DeleteGroupServerAccess(groupName, "192.168.1.105", "22", "admin")
		if err != nil {
			t.Errorf("Unable to delete test server access 2: %s", err)
		}
		err = testutils.DeleteAccount(accountName)
		if err != nil {
			t.Errorf("Unable to delete test account: %s", err)
		}
		err = testutils.DeleteGroup(groupName)
		if err != nil {
			t.Errorf("Unable to delete test group: %s", err)
		}
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create multiple guest accesses
			{
				Config: testAccGroupGuestAccessResourceConfigMultiple(groupName, accountName, "192.168.1.104", "192.168.1.105"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bastion_group_guest_access.test1",
						tfjsonpath.New("ip"),
						knownvalue.StringExact("192.168.1.104"),
					),
					statecheck.ExpectKnownValue(
						"bastion_group_guest_access.test2",
						tfjsonpath.New("ip"),
						knownvalue.StringExact("192.168.1.105"),
					),
				},
			},
		},
	})
}

func testAccGroupGuestAccessResourceConfig(group, account, ip, port, user, comment, ttl string) string { // nolint:unparam
	commentStr := ""
	if comment != "" {
		commentStr = fmt.Sprintf(`comment = "%s"`, comment)
	}
	ttlStr := ""
	if ttl != "" {
		ttlStr = fmt.Sprintf(`ttl = %s`, ttl)
	}

	return providerConfig + fmt.Sprintf(`
resource "bastion_group_guest_access" "test" {
  group   = "%s"
  account = "%s"
  ip      = "%s"
  port    = "%s"
  user    = "%s"
  %s
  %s
}
`, group, account, ip, port, user, commentStr, ttlStr)
}

func testAccGroupGuestAccessResourceConfigWithProtocol(group, account, ip, port, protocol, comment string) string {
	commentStr := ""
	if comment != "" {
		commentStr = fmt.Sprintf(`comment = "%s"`, comment)
	}

	return providerConfig + fmt.Sprintf(`
resource "bastion_group_guest_access" "test" {
  group    = "%s"
  account  = "%s"
  ip       = "%s"
  port     = "%s"
  protocol = "%s"
  %s
}
`, group, account, ip, port, protocol, commentStr)
}

func testAccGroupGuestAccessResourceConfigMultiple(group, account, ip1, ip2 string) string {
	return providerConfig + fmt.Sprintf(`
resource "bastion_group_guest_access" "test1" {
  group   = "%s"
  account = "%s"
  ip      = "%s"
  port    = "22"
  user    = "root"
}

resource "bastion_group_guest_access" "test2" {
  group   = "%s"
  account = "%s"
  ip      = "%s"
  port    = "22"
  user    = "admin"
}
`, group, account, ip1, group, account, ip2)
}
