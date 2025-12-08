# Guest accesses can be imported using the format: group:account:ip:port:user
terraform import bastion_group_guest_access.example kryptonians:kal-el:192.168.1.100:22:postgres

# With protocol
terraform import bastion_group_guest_access.example kryptonians:kal-el:192.168.1.100:22:postgres:sftp

# With protocol and remote_port (protocol must be "portforward")
terraform import bastion_group_guest_access.example kryptonians:kal-el:192.168.1.100:22:postgres:portforward:8080

# With proxy
terraform import bastion_group_guest_access.example kryptonians:kal-el:192.168.1.100:22:postgres:10.0.0.1:22:proxy-user

# With protocol and proxy
terraform import bastion_group_guest_access.example kryptonians:kal-el:192.168.1.100:22:postgres:sftp:10.0.0.1:22:proxy-user

# With protocol, remote_port and proxy (protocol must be "portforward")
terraform import bastion_group_guest_access.example kryptonians:kal-el:192.168.1.100:22:postgres:portforward:8080:10.0.0.1:22:proxy-user

# IPv6 addresses must be wrapped in brackets
terraform import bastion_group_guest_access.example kryptonians:kal-el:[2001:db8::1]:22:root
