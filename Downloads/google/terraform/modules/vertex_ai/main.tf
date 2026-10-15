###############################################################################
# Feature Store.
###############################################################################

resource "google_vertex_ai_featurestore" "platform" {
  name         = "finguard_${var.environment}_featurestore"
  region       = var.region
  labels       = { env = var.environment }

  online_serving_config {
    fixed_node_count = var.environment == "prod" ? 5 : 1
  }

  encryption_spec {
    kms_key_name = var.vertex_kms_key_id
  }

  force_destroy = var.environment != "prod"
}

resource "google_vertex_ai_featurestore_entitytype" "customer" {
  name         = "customer"
  featurestore = google_vertex_ai_featurestore.platform.id
  labels       = { domain = "customer" }
}

resource "google_vertex_ai_featurestore_entitytype" "transaction" {
  name         = "transaction"
  featurestore = google_vertex_ai_featurestore.platform.id
  labels       = { domain = "transaction" }
}

resource "google_vertex_ai_featurestore_entitytype" "device" {
  name         = "device"
  featurestore = google_vertex_ai_featurestore.platform.id
  labels       = { domain = "device" }
}

###############################################################################
# Endpoint.
#
# The real fraud model is uploaded/registered by the CT pipeline; this
# module creates the Endpoint shell so downstream Cloud Run services can
# resolve a target endpoint URL. The CT pipeline uploads the model and
# creates a deployment on this endpoint.
###############################################################################

resource "google_vertex_ai_endpoint" "fraud" {
  name         = "fraud-endpoint-${var.environment}"
  location     = var.region
  display_name = "Fraud Endpoint (${var.environment})"
  labels       = { env = var.environment }

  encryption_spec {
    kms_key_name = var.vertex_kms_key_id
  }
}

###############################################################################
# Model Monitoring — feature/score drift detection on the endpoint.
#
# NOTE: As of google-beta provider 5.45, no resource exists for creating a
# Vertex AI Model Monitoring job. The CT pipeline creates the monitor via
# the Vertex AI Python SDK (`aiplatform.ModelMonitoringJob.create(...)`) and
# configures email alerts in code. See `ml/pipelines/components/deploy.py`.
# Once the provider adds this resource, it can be re-introduced here.
###############################################################################
