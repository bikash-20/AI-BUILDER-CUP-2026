###############################################################################
# BigQuery dataset for log analytics.
###############################################################################

resource "google_bigquery_dataset" "logs" {
  dataset_id  = "finguard_${var.environment}_logs"
  description = "FinGuard AI platform logs for ${var.environment}."
  location    = var.region

  default_table_expiration_ms = 365 * 24 * 60 * 60 * 1000 # 365 days

  labels = {
    env        = var.environment
    managed_by = "terraform"
  }
}

###############################################################################
# Log sinks.
###############################################################################

# Audit logs → BigQuery for compliance analytics.
resource "google_logging_project_sink" "audit_to_bq" {
  name        = "finguard-${var.environment}-audit-to-bq"
  description = "Forward Cloud Audit Logs to BigQuery."

  destination = "bigquery.googleapis.com/projects/${var.project_id}/datasets/${google_bigquery_dataset.logs.dataset_id}"

  filter = <<-EOT
    logName:"cloudaudit.googleapis.com"
  EOT

  unique_writer_identity = true
}

# VPC flow logs → Pub/Sub for streaming into Dataflow.
resource "google_logging_project_sink" "vpc_flow_to_pubsub" {
  name        = "finguard-${var.environment}-vpc-flow-to-pubsub"
  description = "Stream VPC flow logs to Pub/Sub."

  destination = "pubsub.googleapis.com/projects/${var.project_id}/topics/${google_pubsub_topic.vpc_flow.name}"

  filter = <<-EOT
    logName:"vpc_flows"
    severity>=INFO
  EOT

  unique_writer_identity = true
}

###############################################################################
# Pub/Sub topic for VPC flow log streaming.
###############################################################################

resource "google_pubsub_topic" "vpc_flow" {
  name    = "finguard-${var.environment}-vpc-flow"
  labels  = { env = var.environment }
  message_retention_duration = "604800s" # 7 days
}

# Allow the log sink's writer identity to publish to the topic.
resource "google_pubsub_topic_iam_member" "sink_publisher" {
  topic  = google_pubsub_topic.vpc_flow.name
  role   = "roles/pubsub.publisher"
  member = google_logging_project_sink.vpc_flow_to_pubsub.writer_identity
}

###############################################################################
# Log-based metric for fraud decisions.
###############################################################################

resource "google_logging_metric" "fraud_decisions" {
  name        = "finguard_${var.environment}_fraud_decisions"
  description = "Count of fraud decisions emitted by fraud-service."

  filter = <<-EOT
    resource.type="cloud_run_revision"
    resource.labels.service_name="fraud-service"
    jsonPayload.decision:* AND jsonPayload.score:*
  EOT

  metric_descriptor {
    metric_kind = "DELTA"
    value_type  = "INT64"
    labels {
      key         = "decision"
      value_type  = "STRING"
      description = "APPROVE / REVIEW / BLOCK"
    }
    labels {
      key         = "model"
      value_type  = "STRING"
      description = "Model version that produced the decision"
    }
  }
}
