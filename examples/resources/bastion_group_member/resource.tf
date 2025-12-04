
resource "bastion_group_member" "example" {
  group   = "example-group"
  account = "example-account"
}