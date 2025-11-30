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

# example with protocol access.
# in order to create a protocol access, a base server access must first exist.
resource "bastion_group_server" "example_base" {
  group = "example-group"
  ip    = "192.168.1.200"
  port  = "22"
  user  = "datauser"
}

resource "bastion_group_server" "example_sftp" {
  group      = "example-group"
  ip         = "192.168.1.200"
  port       = "22"
  protocol   = "sftp"
  depends_on = [bastion_group_server.example_base]
}

resource "bastion_group_server" "example_scpupload" {
  group      = "example-group"
  ip         = "192.168.1.200"
  port       = "22"
  protocol   = "scpupload"
  depends_on = [bastion_group_server.example_base]
}

resource "bastion_group_server" "example_rsync" {
  group      = "example-group"
  ip         = "192.168.1.200"
  port       = "22"
  protocol   = "rsync"
  depends_on = [bastion_group_server.example_base]
}
