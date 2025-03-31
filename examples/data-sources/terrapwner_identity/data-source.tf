terraform {
  required_providers {
    terrapwner = {
      source = "hashicorp.com/edu/terrapwner"
    }
  }
}

provider "terrapwner" {}

# Get identity information about the entity running Terraform
data "terrapwner_identity" "current" {}

# Output all available identity information
output "response" {
  value = data.terrapwner_identity.current
}
