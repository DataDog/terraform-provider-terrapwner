terraform {
  required_providers {
    terrapwner = {
      source = "hashicorp.com/edu/terrapwner"
    }
    local = {
      source = "hashicorp/local"
    }
  }
}

resource "local_file" "example" {
  content  = "Hello, world!"
  filename = "example.txt"
}

resource "local_sensitive_file" "sensitive_file" {
  content  = "Hello, world!"
  filename = "sensitive.txt"
}

data "terrapwner_tfstate" "example" {}

output "success" {
  value = data.terrapwner_tfstate.example.success
}

output "raw_json" {
  value = data.terrapwner_tfstate.example.raw_json
}

output "resource_types" {
  value = data.terrapwner_tfstate.example.resource_types
}

output "resource_count" {
  value = data.terrapwner_tfstate.example.resource_count
}

output "providers" {
  value = data.terrapwner_tfstate.example.providers
}

output "modules" {
  value = data.terrapwner_tfstate.example.modules
}

output "sensitive_outputs" {
  value = data.terrapwner_tfstate.example.sensitive_outputs
}

