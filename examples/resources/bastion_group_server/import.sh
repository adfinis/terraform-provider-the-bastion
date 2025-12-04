
# Expected format: group:ip:port:user or group:ip:port:user:proxy_ip:proxy_port:proxy_user

terraform import bastion_group_server.example 'example-group:192.168.1.100:22:ssh-user'

terraform import bastion_group_server.example2 'example-group:10.0.0.0/24:*:*'

terraform import bastion_group_server.example3 'example-group:[2001:db8::1]:22:root'

terraform import bastion_group_server.example4 'example-group:172.16.0.50:22:app:10.0.0.1:22:proxy_user'

terraform import bastion_group_server.example5 'example-group:[2001:db8::1]:22:admin:[fd00::1]:22:proxy_user'
