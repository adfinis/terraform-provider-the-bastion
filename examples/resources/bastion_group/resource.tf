
resource "bastion_group" "example" {
  group             = "example-group"
  owner             = "bastionadmin"
  key_algo          = "ed25519"
  mfa_required      = "totp"
  idle_lock_timeout = 900
  idle_kill_timeout = 1800
  guest_ttl_limit   = 86400
}
