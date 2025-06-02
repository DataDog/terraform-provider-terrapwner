terraform {
  required_providers {
    terrapwner = {
      source = "hashicorp.com/DataDog/terrapwner"
    }
  }
}

data "terrapwner_env_dump" "current" {}

output "response" {
  value = data.terrapwner_env_dump.current.vars
}
