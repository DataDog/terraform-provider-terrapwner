terraform {
  required_providers {
    terrapwner = {
      source = "hashicorp.com/DataDog/terrapwner"
    }
  }
}

provider "terrapwner" {}

data "terrapwner_env_dump" "current" {}

output "env" {
  value = data.terrapwner_env_dump.current.vars["PWD"]
}
