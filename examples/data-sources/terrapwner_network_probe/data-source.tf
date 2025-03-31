terraform {
  required_providers {
    terrapwner = {
      source = "hashicorp.com/edu/terrapwner"
    }
  }
}

provider "terrapwner" {}

# Probe DNS resolution to example.com
data "terrapwner_network_probe" "dns" {
  type    = "dns"
  host    = "example.com"
  timeout = 5 # 5 seconds timeout
}

# Probe TCP connection to example.com:80
data "terrapwner_network_probe" "tcp" {
  type    = "tcp"
  host    = "example.com"
  port    = 80
  timeout = 10 # 10 seconds timeout
}

# Probe UDP connection to example.com:53
data "terrapwner_network_probe" "udp" {
  type = "udp"
  host = "example.com"
  port = 53
  # Using default timeout (5 seconds)
}

# Probe ICMP ping to example.com
data "terrapwner_network_probe" "icmp" {
  type    = "icmp"
  host    = "example.com"
  timeout = 3 # 3 seconds timeout
}

# Output complete DNS probe response
output "dns_response" {
  value = data.terrapwner_network_probe.dns
}

# Output complete TCP probe response
output "tcp_response" {
  value = data.terrapwner_network_probe.tcp
}

# Output complete UDP probe response
output "udp_response" {
  value = data.terrapwner_network_probe.udp
}

# Output complete ICMP probe response
output "icmp_response" {
  value = data.terrapwner_network_probe.icmp
}
