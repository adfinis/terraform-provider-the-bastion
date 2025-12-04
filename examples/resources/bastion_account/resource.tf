
resource "bastion_account" "example" {
  account           = "cal-el"
  uid_auto          = true
  public_key        = file("id_ed25519.pub")
  max_inactive_days = 90
  mfa_totp_required = "yes"
}