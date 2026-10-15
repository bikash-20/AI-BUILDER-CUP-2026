###############################################################################
# Folder hierarchy under the organization root.
# `var.org_id` should be the numeric org ID (e.g. "123456789012").
###############################################################################

resource "google_folder" "finguard" {
  display_name = "FinGuard AI"
  parent       = "organizations/${var.org_id}"

  # Do not delete when running `terraform destroy`; use a flag instead.
  lifecycle {
    prevent_destroy = true
  }
}

resource "google_folder" "environments" {
  display_name = "Environments"
  parent       = google_folder.finguard.id
}

resource "google_folder" "dev" {
  display_name = "Dev"
  parent       = google_folder.environments.id
}

resource "google_folder" "staging" {
  display_name = "Staging"
  parent       = google_folder.environments.id
}

resource "google_folder" "prod" {
  display_name = "Prod"
  parent       = google_folder.environments.id
}

###############################################################################
# Baseline org policies — these enforce platform-wide guardrails.
###############################################################################

# Restrict allowed VM external IPs.
resource "google_organization_policy" "vm_external_ip" {
  org_id     = var.org_id
  constraint = "constraints/compute.vmExternalIpAccess"

  list_policy {
    deny {
      all = true
    }
  }
}

# Require OS Login for all VMs.
resource "google_organization_policy" "os_login" {
  org_id     = var.org_id
  constraint = "constraints/compute.requireOsLogin"

  boolean_policy {
    enforced = true
  }
}

# Restrict service account key creation (use workload identity instead).
resource "google_organization_policy" "disable_sa_key_creation" {
  org_id     = var.org_id
  constraint = "constraints/iam.disableServiceAccountKeyCreation"

  boolean_policy {
    enforced = true
  }
}

# Restrict allowed locations for new resources (data residency).
resource "google_organization_policy" "allowed_locations" {
  org_id     = var.org_id
  constraint = "constraints/gcp.resourceLocations"

  list_policy {
    allow {
      values = var.allowed_regions
    }
  }
}
