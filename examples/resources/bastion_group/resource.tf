
resource "bastion_group" "example" {
  group             = "example-group"
  owner             = "bastionadmin"
  key_algo          = "ed25519"
  mfa_required      = "totp"
  idle_lock_timeout = "2h"
  idle_kill_timeout = "6h"
  guest_ttl_limit   = "7d"
}
