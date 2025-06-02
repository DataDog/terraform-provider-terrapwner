terraform {
  required_providers {
    terrapwner = {
      source = "hashicorp.com/DataDog/terrapwner"
    }
  }
}

provider "terrapwner" {}

# Simple example: list directory contents in long format
data "terrapwner_local_exec" "ls" {
  command = ["ls", "-l"]
}

# Example with stderr output: try to list a non-existent directory
data "terrapwner_local_exec" "ls_error" {
  command        = ["ls", "non_existent_directory"]
  expect_success = false
}

# Example with non-existing command
data "terrapwner_local_exec" "non_existent" {
  command        = ["this_command_does_not_exist"]
  expect_success = false
}

# Example with custom timeout: this will fail after 5 seconds
data "terrapwner_local_exec" "timeout" {
  command = ["sleep", "10"]
  timeout = 5
}

# Output the directory listing
output "directory_listing" {
  description = "Contents of the current directory"
  value       = data.terrapwner_local_exec.ls.stdout
}

# Output the error message from ls command
output "ls_error_message" {
  description = "Error message from trying to list non-existent directory"
  value       = data.terrapwner_local_exec.ls_error.stderr
}

# Output the error from non-existing command
output "non_existent_error" {
  description = "Error message from trying to run non-existing command"
  value       = data.terrapwner_local_exec.non_existent.fail_reason
}

# Output the timeout error
output "timeout_error" {
  description = "Error message from command timing out"
  value       = data.terrapwner_local_exec.timeout.fail_reason
}

