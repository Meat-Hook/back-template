datacenter = "DC"
data_dir = "/opt/nomad/"
bind_addr = "X.X.X.X"
region = "global"

server {
  enabled = true
  bootstrap_expect = 3
}

client {
  enabled = true
}

consul {
  address = "X.X.X.X:8500"
  server_service_name = "nomad"
  client_service_name = "nomad-client"
  auto_advertise = true
  server_auto_join = true
  client_auto_join = true
}
