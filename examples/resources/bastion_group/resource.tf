
resource "bastion_group" "example" {
  group    = "example-group"
  owner    = "bastionadmin"
  key_algo = "ed25519"
}