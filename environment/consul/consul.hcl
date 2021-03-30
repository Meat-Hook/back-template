datacenter = "DC"
data_dir = "/opt/consul/"
encrypt = "CONSUL_KEY"
ca_file = "/etc/consul.d/consul-agent-ca.pem"
cert_file = "/etc/consul.d/dc-server-consul-X.pem"
key_file = "/etc/consul.d/dc-server-consul-X-key.pem"
verify_incoming = true
verify_outgoing = true
verify_server_hostname = true

bind_addr = "X.X.X.X"

retry_join = ["X.X.X.X", "X.X.X.X"]

acl = {
  enabled = false
  default_policy = "allow"
  enable_token_persistence = true
}

performance {
  raft_multiplier = 1
}
