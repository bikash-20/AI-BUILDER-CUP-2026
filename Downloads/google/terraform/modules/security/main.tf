###############################################################################
# KMS — one keyring per service family, rotation enabled, Envelope CMEK.
###############################################################################

resource "google_kms_key_ring" "platform" {
  name     = "finguard-${var.environment}-keyring"
  location = var.primary_region
}

resource "google_kms_crypto_key" "bigquery" {
  name            = "bigquery"
  key_ring        = google_kms_key_ring.platform.id
  rotation_period = "7776000s" # 90 days

  version_template {
    algorithm = "GOOGLE_SYMMETRIC_ENCRYPTION"
  }

  lifecycle {
    prevent_destroy = true
  }
}

resource "google_kms_crypto_key" "gcs" {
  name            = "gcs"
  key_ring        = google_kms_key_ring.platform.id
  rotation_period = "7776000s"

  version_template {
    algorithm = "GOOGLE_SYMMETRIC_ENCRYPTION"
  }

  lifecycle {
    prevent_destroy = true
  }
}

resource "google_kms_crypto_key" "pubsub" {
  name            = "pubsub"
  key_ring        = google_kms_key_ring.platform.id
  rotation_period = "7776000s"

  version_template {
    algorithm = "GOOGLE_SYMMETRIC_ENCRYPTION"
  }

  lifecycle {
    prevent_destroy = true
  }
}

resource "google_kms_crypto_key" "alloydb" {
  name            = "alloydb"
  key_ring        = google_kms_key_ring.platform.id
  rotation_period = "7776000s"

  version_template {
    algorithm = "GOOGLE_SYMMETRIC_ENCRYPTION"
  }

  lifecycle {
    prevent_destroy = true
  }
}

resource "google_kms_crypto_key" "vertex" {
  name            = "vertex"
  key_ring        = google_kms_key_ring.platform.id
  rotation_period = "7776000s"

  version_template {
    algorithm = "GOOGLE_SYMMETRIC_ENCRYPTION"
  }

  lifecycle {
    prevent_destroy = true
  }
}

###############################################################################
# IAM — service accounts for platform services. Bindings to keys happen
# in the per-service module to keep this module focused on primitives.
###############################################################################

resource "google_service_account" "fraud_runtime" {
  account_id   = "fraud-runtime-${var.environment}"
  display_name = "Fraud Service Runtime"
}

resource "google_service_account" "risk_runtime" {
  account_id   = "risk-runtime-${var.environment}"
  display_name = "Risk Service Runtime"
}

resource "google_service_account" "chat_runtime" {
  account_id   = "chat-runtime-${var.environment}"
  display_name = "Chat BFF Runtime"
}

resource "google_service_account" "doc_runtime" {
  account_id   = "doc-runtime-${var.environment}"
  display_name = "Document Service Runtime"
}

resource "google_service_account" "pipeline_runner" {
  account_id   = "pipeline-runner-${var.environment}"
  display_name = "Vertex AI Pipeline Runner"
}

# Allow pipeline runner to impersonate runtime service accounts
# (workload identity binding pattern).
resource "google_service_account_iam_member" "pipeline_runner_actas" {
  for_each = toset([
    google_service_account.fraud_runtime.email,
    google_service_account.risk_runtime.email,
    google_service_account.chat_runtime.email,
    google_service_account.doc_runtime.email,
  ])

  service_account_id = each.value
  role               = "roles/iam.serviceAccountTokenCreator"
  member             = "serviceAccount:${google_service_account.pipeline_runner.email}"
}

###############################################################################
# VPC Service Controls perimeter.
###############################################################################

resource "google_access_context_manager_service_perimeter" "platform" {
  parent = "accessPolicies/${var.access_policy_id}"
  name   = "accessPolicies/${var.access_policy_id}/servicePerimeters/finguard_${var.environment}"
  title  = "FinGuard ${var.environment} Perimeter"

  status {
    restricted_services = [
      "bigquery.googleapis.com",
      "storage.googleapis.com",
      "pubsub.googleapis.com",
      "aiplatform.googleapis.com",
      "run.googleapis.com",
      "documentai.googleapis.com",
    ]

    resources = [
      "projects/${var.project_id}",
    ]

    vpc_accessible_services {
      enable_restriction = true
      allowed_services = [
        "bigquery.googleapis.com",
        "storage.googleapis.com",
        "pubsub.googleapis.com",
        "aiplatform.googleapis.com",
        "run.googleapis.com",
        "documentai.googleapis.com",
      ]
    }
  }
}

###############################################################################
# Cloud Armor — managed WAF + rate limit for the public API.
###############################################################################

resource "google_compute_security_policy" "platform" {
  name        = "finguard-${var.environment}-waf"
  description = "WAF + rate limit policy for public FinGuard AI endpoints."

  # Adaptive Protection — auto-mitigates L7 attacks.
  adaptive_protection_config {
    layer_7_ddos_defense_config {
      enable = true
    }
  }
}

resource "google_compute_security_policy_rule" "sqli" {
  security_policy = google_compute_security_policy.platform.id
  priority        = 1000
  description     = "SQLi protection (managed)"

  match {
    expr {
      expression = "evaluatePreconfiguredWaf('sqli-v33-stable')"
    }
  }

  action = "deny(403)"
}

resource "google_compute_security_policy_rule" "xss" {
  security_policy = google_compute_security_policy.platform.id
  priority        = 1001
  description     = "XSS protection (managed)"

  match {
    expr {
      expression = "evaluatePreconfiguredWaf('xss-v33-stable')"
    }
  }

  action = "deny(403)"
}

resource "google_compute_security_policy_rule" "rate_limit" {
  security_policy = google_compute_security_policy.platform.id
  priority        = 2000
  description     = "Rate limit per source IP"

  match {
    versioned_expr = "SRC_IPS_V1"
  }

  action = "throttle"
  rate_limit_options {
    conform_action      = "allow"
    exceed_action       = "deny(429)"
    enforce_on_key      = "IP"
    rate_limit_threshold {
      count        = 600
      interval_sec = 60
    }
  }
}
