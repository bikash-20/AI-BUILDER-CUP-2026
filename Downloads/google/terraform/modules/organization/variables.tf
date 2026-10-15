variable "org_id" {
  description = "Numeric Google Cloud organization ID (e.g. 123456789012)."
  type        = string
}

variable "allowed_regions" {
  description = "Allowed regions for resource creation. Enforced by gcp.resourceLocations policy."
  type        = list(string)
  default = [
    "us-central1",
    "us-east1",
    "europe-west1",
  ]
}
