
resource "bastion_group_guest_access" "guest_db_access" {
  group   = "kryptonians"
  account = "jonnjonzz"
  ip      = "192.168.1.100"
  port    = "22"
  user    = "*"
}
