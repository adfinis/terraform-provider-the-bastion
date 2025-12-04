# basic example
resource "bastion_group_server" "example" {
  group = "example-group"
  ip    = "192.168.1.100"
  port  = "22"
  user  = "ssh-user"
}

# example with an ssh proxyjump
resource "bastion_group_server" "example_proxy" {
  group      = "example-group"
  ip         = "192.168.1.100"
  port       = "22"
  user       = "ssh-user"
  proxy_ip   = "10.10.10.10"
  proxy_port = "22"
  proxy_user = "proxyuser"
}
