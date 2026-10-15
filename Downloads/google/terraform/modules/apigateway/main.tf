###############################################################################
# API & API Config.
###############################################################################

resource "google_api_gateway_api" "platform" {
  provider     = google-beta
  api_id       = "finguard-${var.environment}"
  display_name = "FinGuard AI (${var.environment})"
}

resource "google_api_gateway_api_config" "platform" {
  provider             = google-beta
  api                  = google_api_gateway_api.platform.api_id
  api_config_id_prefix = "finguard-${var.environment}-"

  openapi_documents {
    document {
      path     = "openapi.yaml"
      contents = filebase64("${path.module}/openapi.yaml")
    }
  }

  gateway_config {
    backend_config {
      google_service_account = var.gateway_runtime_sa
    }
  }

  lifecycle {
    create_before_destroy = true
  }
}

###############################################################################
# Gateway itself.
###############################################################################

resource "google_api_gateway_gateway" "platform" {
  provider     = google-beta
  api_config   = google_api_gateway_api_config.platform.id
  gateway_id   = "finguard-${var.environment}"
  display_name = "FinGuard AI Gateway (${var.environment})"
  region       = var.region
}
