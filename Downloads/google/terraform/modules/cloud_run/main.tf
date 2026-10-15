###############################################################################
# Cloud Run VPC connector (for direct VPC egress).
###############################################################################

resource "google_vpc_access_connector" "connector" {
  name          = "finguard-${var.environment}-connector"
  region        = var.region
  network       = var.network_id
  ip_cidr_range = "10.50.0.0/28"

  min_instances = 2
  max_instances = 10
}

###############################################################################
# Service map. Adding a service means adding a key to this map and a
# corresponding local in the for_each loop below.
###############################################################################

locals {
  services = {
    fraud-service = {
      image          = var.fraud_image
      service_account = var.fraud_runtime_sa
      port            = 8080
      env = {
        GCP_PROJECT_ID  = var.project_id
        ENVIRONMENT     = var.environment
        VERTEX_ENDPOINT = var.vertex_fraud_endpoint
        BIGQUERY_DATASET = "finguard_${var.environment}"
        FEATURE_STORE_ID = var.feature_store_id
      }
      min_instances = var.environment == "prod" ? 2 : 0
      max_instances = var.environment == "prod" ? 50 : 5
      cpu           = "1"
      memory        = "512Mi"
      concurrency   = 80
    }

    risk-service = {
      image          = var.risk_image
      service_account = var.risk_runtime_sa
      port            = 8080
      env = {
        GCP_PROJECT_ID   = var.project_id
        ENVIRONMENT      = var.environment
        BIGQUERY_DATASET = "finguard_${var.environment}"
      }
      min_instances = 1
      max_instances = 20
      cpu           = "1"
      memory        = "1Gi"
      concurrency   = 40
    }

    chat-bff = {
      image          = var.chat_bff_image
      service_account = var.chat_runtime_sa
      port            = 8080
      env = {
        GCP_PROJECT_ID  = var.project_id
        ENVIRONMENT     = var.environment
        GEMINI_SERVICE_URL = var.gemini_service_url
      }
      min_instances = 1
      max_instances = 30
      cpu           = "1"
      memory        = "512Mi"
      concurrency   = 80
    }

    document-service = {
      image          = var.doc_image
      service_account = var.doc_runtime_sa
      port            = 8080
      env = {
        GCP_PROJECT_ID  = var.project_id
        ENVIRONMENT     = var.environment
        DOCAI_PROCESSOR_ID = var.docai_processor_id
      }
      min_instances = 0
      max_instances = 10
      cpu           = "1"
      memory        = "1Gi"
      concurrency   = 10
    }
  }
}

resource "google_cloud_run_v2_service" "services" {
  for_each = local.services

  name     = "finguard-${var.environment}-${each.key}"
  location = var.region

  template {
    service_account = each.value.service_account
    scaling {
      min_instance_count = each.value.min_instances
      max_instance_count = each.value.max_instances
    }

    containers {
      image = each.value.image
      ports {
        container_port = each.value.port
      }

      dynamic "env" {
        for_each = each.value.env
        content {
          name  = env.key
          value = env.value
        }
      }

      resources {
        cpu_idle = true
        limits = {
          cpu    = each.value.cpu
          memory = each.value.memory
        }
      }
    }

    vpc_access {
      connector = google_vpc_access_connector.connector.id
      egress    = "PRIVATE_RANGES_ONLY"
    }

    max_instance_request_concurrency = each.value.concurrency
  }

  traffic {
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
    percent = 100
  }

  labels = {
    env        = var.environment
    managed_by = "terraform"
  }

  lifecycle {
    ignore_changes = [
      template[0].containers[0].image,
      client,
      client_version,
    ]
  }
}

###############################################################################
# Public invoker role (api-gateway) — but services are NOT public; they
# are only invoked via API Gateway + IAP. Deny all-unauthenticated access.
###############################################################################

resource "google_cloud_run_v2_service_iam_member" "no_public_invoker" {
  for_each = local.services

  project  = var.project_id
  location = var.region
  name     = google_cloud_run_v2_service.services[each.key].name
  role     = "roles/run.invoker"
  member   = "allUsers"

  # The `condition` block makes this an EXPLICIT DENY.
  condition {
    title       = "deny-unauthenticated"
    description = "No unauthenticated access; only IAP-protected gateway."
    expression  = "request.auth.claims == null && request.auth.claims.email_verified == false"
  }
}

# Allow the API Gateway runtime SA to invoke these services.
resource "google_cloud_run_v2_service_iam_member" "gateway_invoker" {
  for_each = local.services

  project  = var.project_id
  location = var.region
  name     = google_cloud_run_v2_service.services[each.key].name
  role     = "roles/run.invoker"
  member   = "serviceAccount:${var.gateway_runtime_sa}"
}
