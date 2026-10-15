###############################################################################
# Dev environment.
# Composes the platform modules into a single project.
###############################################################################

provider "google" {
  project = var.project_id
  region  = var.region
}

provider "google-beta" {
  project = var.project_id
  region  = var.region
}

locals {
  env     = "dev"
  region  = "us-central1"
  dr_region = "us-east1"
}

###############################################################################
# Network.
###############################################################################

module "network" {
  source = "../../modules/network"

  environment    = local.env
  primary_region = local.region
  dr_region      = local.dr_region
}

###############################################################################
# Security (CMEK + VPC-SC + Cloud Armor + service accounts).
###############################################################################

module "security" {
  source = "../../modules/security"

  project_id      = var.project_id
  environment     = local.env
  primary_region  = local.region
  access_policy_id = var.access_policy_id
}

###############################################################################
# Data (BigQuery, Pub/Sub, GCS).
###############################################################################

module "data" {
  source = "../../modules/data"

  project_id          = var.project_id
  environment         = local.env
  region              = local.region
  bigquery_kms_key_id = module.security.bigquery_key_id
  gcs_kms_key_id      = module.security.gcs_key_id
  pubsub_kms_key_id   = module.security.pubsub_key_id
}

###############################################################################
# Vertex AI (Featurestore + Endpoint).
###############################################################################

module "vertex_ai" {
  source = "../../modules/vertex_ai"

  project_id        = var.project_id
  environment       = local.env
  region            = local.region
  vertex_kms_key_id = module.security.vertex_key_id
  alert_email       = var.alert_email
}

###############################################################################
# Cloud Run services (shell only — images come from Cloud Deploy).
###############################################################################

module "cloud_run" {
  source = "../../modules/cloud_run"

  project_id        = var.project_id
  environment       = local.env
  region            = local.region
  network_id        = module.network.vpc_self_link
  fraud_runtime_sa  = module.security.fraud_runtime_sa
  risk_runtime_sa   = module.security.risk_runtime_sa
  chat_runtime_sa   = module.security.chat_runtime_sa
  doc_runtime_sa    = module.security.doc_runtime_sa
  gateway_runtime_sa = module.security.pipeline_runner_sa # placeholder
  vertex_fraud_endpoint = module.vertex_ai.fraud_endpoint_resource_name
  feature_store_id  = module.vertex_ai.featurestore_id
  gemini_service_url = ""
  docai_processor_id = ""
}

###############################################################################
# API Gateway.
###############################################################################

module "apigateway" {
  source = "../../modules/apigateway"

  project_id        = var.project_id
  environment       = local.env
  region            = local.region
  gateway_runtime_sa = module.security.pipeline_runner_sa
  fraud_service_url = module.cloud_run.service_urls["fraud-service"]
  risk_service_url  = module.cloud_run.service_urls["risk-service"]
  chat_bff_url      = module.cloud_run.service_urls["chat-bff"]
  doc_service_url   = module.cloud_run.service_urls["document-service"]
}

###############################################################################
# Monitoring (SLOs, alerts, dashboards).
###############################################################################

module "monitoring" {
  source = "../../modules/monitoring"

  project_id  = var.project_id
  environment = local.env
  alert_email = var.alert_email
}
