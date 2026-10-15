variable "environment" {
  description = "Environment name (dev / staging / prod). Used for naming."
  type        = string
}

variable "primary_region" {
  description = "Primary GCP region."
  type        = string
  default     = "us-central1"
}

variable "dr_region" {
  description = "Disaster recovery GCP region."
  type        = string
  default     = "us-east1"
}

variable "primary_cidr" {
  description = "Primary subnet CIDR."
  type        = string
  default     = "10.10.0.0/20"
}

variable "dr_cidr" {
  description = "DR subnet CIDR."
  type        = string
  default     = "10.11.0.0/20"
}
