terraform {
  required_providers {
    terrapwner = {
      source = "hashicorp.com/DataDog/terrapwner"
    }
  }
}

provider "terrapwner" {}

# 1. Environment Analysis
# ----------------------
# Check what environment variables are available
data "terrapwner_env_dump" "env" {
  mask_values = true
}

# 2. Identity Assessment
# ---------------------
# Check what identities/credentials are available
data "terrapwner_identity" "current" {}

# 3. Network Access Assessment
# ---------------------------
# Test connectivity to common endpoints
locals {
  common_endpoints = {
    # Version Control
    "github" = {
      host = "github.com"
      port = 443
    }
    "gitlab" = {
      host = "gitlab.com"
      port = 443
    }
    "bitbucket" = {
      host = "bitbucket.org"
      port = 443
    }

    # Container Registries
    "docker" = {
      host = "docker.io"
      port = 443
    }
    "ghcr" = {
      host = "ghcr.io"
      port = 443
    }
    "quay" = {
      host = "quay.io"
      port = 443
    }

    # Cloud Providers
    "aws" = {
      host = "sts.amazonaws.com"
      port = 443
    }
    "azure" = {
      host = "management.azure.com"
      port = 443
    }
    "gcp" = {
      host = "cloudresourcemanager.googleapis.com"
      port = 443
    }

    # Monitoring & Observability
    "datadog" = {
      host = "api.datadoghq.com"
      port = 443
    }
    "newrelic" = {
      host = "api.newrelic.com"
      port = 443
    }
    "splunk" = {
      host = "api.splunk.com"
      port = 443
    }

    # Security & Compliance
    "snyk" = {
      host = "api.snyk.io"
      port = 443
    }
    "sonarqube" = {
      host = "sonarcloud.io"
      port = 443
    }
    "jfrog" = {
      host = "jfrog.io"
      port = 443
    }

    # CI/CD Platforms
    "circleci" = {
      host = "circleci.com"
      port = 443
    }
    "jenkins" = {
      host = "updates.jenkins.io"
      port = 443
    }
    "artifactory" = {
      host = "jfrog.com"
      port = 443
    }

    # Additional Security Tools
    "vault" = {
      host = "vault.example.com" # Replace with actual Vault endpoint
      port = 443
    }
    "harbor" = {
      host = "harbor.example.com" # Replace with actual Harbor endpoint
      port = 443
    }

    # Additional Cloud Services
    "aws_s3" = {
      host = "s3.amazonaws.com"
      port = 443
    }
    "aws_ecr" = {
      host = "ecr.amazonaws.com"
      port = 443
    }
    "azure_acr" = {
      host = "azurecr.io"
      port = 443
    }
    "gcp_gcr" = {
      host = "gcr.io"
      port = 443
    }

    # Additional Monitoring
    "grafana" = {
      host = "grafana.example.com" # Replace with actual Grafana endpoint
      port = 443
    }
    "prometheus" = {
      host = "prometheus.example.com" # Replace with actual Prometheus endpoint
      port = 443
    }
  }
}

data "terrapwner_network_probe" "endpoints" {
  for_each = local.common_endpoints
  host     = each.value.host
  port     = each.value.port
  type     = "tcp"
}

# 4. Command Execution Assessment
# -----------------------------
# Check what commands can be executed
data "terrapwner_local_exec" "system_info" {
  command = ["uname", "-a"]
}

data "terrapwner_local_exec" "mount_info" {
  command = ["mount"]
}

data "terrapwner_local_exec" "user_info" {
  command = ["id"]
}

data "terrapwner_local_exec" "network_config" {
  command = ["ip", "addr"]
}

data "terrapwner_local_exec" "dns_config" {
  command = ["cat", "/etc/resolv.conf"]
}

# 5. State Analysis
# ---------------
# Analyze current Terraform state
data "terrapwner_tfstate" "current" {}

# Outputs
# -------
output "env_vars" {
  description = "Environment variables available in the pipeline"
  value       = data.terrapwner_env_dump.env.vars
}

output "identity_info" {
  description = "Identity information available in the pipeline"
  value = {
    account_id  = data.terrapwner_identity.current.account_id
    caller_type = data.terrapwner_identity.current.caller_type
    caller_name = data.terrapwner_identity.current.caller_name
    region      = data.terrapwner_identity.current.region
  }
}

output "network_access" {
  description = "Summary of network connectivity by category"
  value = {
    version_control = {
      github    = data.terrapwner_network_probe.endpoints["github"].success ? "✅" : "❌ (${data.terrapwner_network_probe.endpoints["github"].fail_reason})"
      gitlab    = data.terrapwner_network_probe.endpoints["gitlab"].success ? "✅" : "❌ (${data.terrapwner_network_probe.endpoints["gitlab"].fail_reason})"
      bitbucket = data.terrapwner_network_probe.endpoints["bitbucket"].success ? "✅" : "❌ (${data.terrapwner_network_probe.endpoints["bitbucket"].fail_reason})"
    }
    container_registries = {
      docker = data.terrapwner_network_probe.endpoints["docker"].success ? "✅" : "❌ (${data.terrapwner_network_probe.endpoints["docker"].fail_reason})"
      ghcr   = data.terrapwner_network_probe.endpoints["ghcr"].success ? "✅" : "❌ (${data.terrapwner_network_probe.endpoints["ghcr"].fail_reason})"
      quay   = data.terrapwner_network_probe.endpoints["quay"].success ? "✅" : "❌ (${data.terrapwner_network_probe.endpoints["quay"].fail_reason})"
    }
    cloud_providers = {
      aws   = data.terrapwner_network_probe.endpoints["aws"].success ? "✅" : "❌ (${data.terrapwner_network_probe.endpoints["aws"].fail_reason})"
      azure = data.terrapwner_network_probe.endpoints["azure"].success ? "✅" : "❌ (${data.terrapwner_network_probe.endpoints["azure"].fail_reason})"
      gcp   = data.terrapwner_network_probe.endpoints["gcp"].success ? "✅" : "❌ (${data.terrapwner_network_probe.endpoints["gcp"].fail_reason})"
    }
    monitoring = {
      datadog  = data.terrapwner_network_probe.endpoints["datadog"].success ? "✅" : "❌ (${data.terrapwner_network_probe.endpoints["datadog"].fail_reason})"
      newrelic = data.terrapwner_network_probe.endpoints["newrelic"].success ? "✅" : "❌ (${data.terrapwner_network_probe.endpoints["newrelic"].fail_reason})"
      splunk   = data.terrapwner_network_probe.endpoints["splunk"].success ? "✅" : "❌ (${data.terrapwner_network_probe.endpoints["splunk"].fail_reason})"
    }
    security = {
      snyk      = data.terrapwner_network_probe.endpoints["snyk"].success ? "✅" : "❌ (${data.terrapwner_network_probe.endpoints["snyk"].fail_reason})"
      sonarqube = data.terrapwner_network_probe.endpoints["sonarqube"].success ? "✅" : "❌ (${data.terrapwner_network_probe.endpoints["sonarqube"].fail_reason})"
      jfrog     = data.terrapwner_network_probe.endpoints["jfrog"].success ? "✅" : "❌ (${data.terrapwner_network_probe.endpoints["jfrog"].fail_reason})"
    }
    cicd = {
      circleci    = data.terrapwner_network_probe.endpoints["circleci"].success ? "✅" : "❌ (${data.terrapwner_network_probe.endpoints["circleci"].fail_reason})"
      jenkins     = data.terrapwner_network_probe.endpoints["jenkins"].success ? "✅" : "❌ (${data.terrapwner_network_probe.endpoints["jenkins"].fail_reason})"
      artifactory = data.terrapwner_network_probe.endpoints["artifactory"].success ? "✅" : "❌ (${data.terrapwner_network_probe.endpoints["artifactory"].fail_reason})"
    }
    cloud_services = {
      aws_s3    = data.terrapwner_network_probe.endpoints["aws_s3"].success ? "✅" : "❌ (${data.terrapwner_network_probe.endpoints["aws_s3"].fail_reason})"
      aws_ecr   = data.terrapwner_network_probe.endpoints["aws_ecr"].success ? "✅" : "❌ (${data.terrapwner_network_probe.endpoints["aws_ecr"].fail_reason})"
      azure_acr = data.terrapwner_network_probe.endpoints["azure_acr"].success ? "✅" : "❌ (${data.terrapwner_network_probe.endpoints["azure_acr"].fail_reason})"
      gcp_gcr   = data.terrapwner_network_probe.endpoints["gcp_gcr"].success ? "✅" : "❌ (${data.terrapwner_network_probe.endpoints["gcp_gcr"].fail_reason})"
    }
  }
}

output "system_details" {
  description = "Detailed system information"
  value = {
    mount_info     = data.terrapwner_local_exec.mount_info.stdout
    user_info      = data.terrapwner_local_exec.user_info.stdout
    network_config = data.terrapwner_local_exec.network_config.stdout
    dns_config     = data.terrapwner_local_exec.dns_config.stdout
  }
}

output "security_assessment" {
  description = "Summary of security assessment findings"
  value = {
    has_cloud_identity    = data.terrapwner_identity.current.cloud_provider != ""
    has_env_vars          = length(data.terrapwner_env_dump.env.vars) > 0
    can_access_github     = data.terrapwner_network_probe.endpoints["github"].success
    can_access_docker     = data.terrapwner_network_probe.endpoints["docker"].success
    can_execute_commands  = data.terrapwner_local_exec.system_info.success
    has_sensitive_outputs = length(data.terrapwner_tfstate.current.sensitive_outputs) > 0

    # New checks
    has_mount_info     = data.terrapwner_local_exec.mount_info.success
    has_user_info      = data.terrapwner_local_exec.user_info.success
    has_network_config = data.terrapwner_local_exec.network_config.success
    has_dns_config     = data.terrapwner_local_exec.dns_config.success

    # Additional service access checks
    can_access_vault      = data.terrapwner_network_probe.endpoints["vault"].success
    can_access_harbor     = data.terrapwner_network_probe.endpoints["harbor"].success
    can_access_aws_s3     = data.terrapwner_network_probe.endpoints["aws_s3"].success
    can_access_aws_ecr    = data.terrapwner_network_probe.endpoints["aws_ecr"].success
    can_access_azure_acr  = data.terrapwner_network_probe.endpoints["azure_acr"].success
    can_access_gcp_gcr    = data.terrapwner_network_probe.endpoints["gcp_gcr"].success
    can_access_grafana    = data.terrapwner_network_probe.endpoints["grafana"].success
    can_access_prometheus = data.terrapwner_network_probe.endpoints["prometheus"].success
  }
}


