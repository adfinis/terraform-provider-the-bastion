
# Expected formats:
# - group:ip:port:user
# - group:ip:port:user:protocol
# - group:ip:port:user:proxy_ip:proxy_port:proxy_user
# - group:ip:port:user:protocol:proxy_ip:proxy_port:proxy_user

terraform import bastion_group_server.example 'kryptonians:192.168.1.100:22:kal-el'

terraform import bastion_group_server.example2 'kryptonians:10.0.0.0/24:*:*'

terraform import bastion_group_server.example3 'kryptonians:[2001:db8::1]:22:kal-el'

terraform import bastion_group_server.example4 'kryptonians:172.16.0.50:22:kal-el:10.0.0.1:22:jor-el'

terraform import bastion_group_server.example5 'kryptonians:[2001:db8::1]:22:kal-el:[fd00::1]:22:jor-el'

terraform import bastion_group_server.example6 'kryptonians:192.168.1.200:22::sftp'

terraform import bastion_group_server.example7 'kryptonians:10.0.0.50:22::rsync:192.168.1.1:22:jor-el'
