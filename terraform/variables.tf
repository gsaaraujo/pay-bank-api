variable "environment" {
  type = string

  validation {
    condition     = contains(["Staging", "Prod"], var.environment)
    error_message = "Environment must be 'Staging' or 'Prod'"
  }
}

variable "db_password" {
  type      = string
  sensitive = true
}

variable "access_token_signing_key" {
  type      = string
  sensitive = true
}

variable "access_key" {
  type      = string
  sensitive = true
}

variable "secret_key" {
  type      = string
  sensitive = true
}
