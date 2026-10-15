###############################################################################
# Notification channels.
###############################################################################

resource "google_monitoring_notification_channel" "email" {
  display_name = "FinGuard ${var.environment} - Email"
  type         = "email"
  labels = {
    email_address = var.alert_email
  }
}

resource "google_monitoring_notification_channel" "pubsub" {
  display_name = "FinGuard ${var.environment} - PagerDuty bridge"
  type         = "pubsub"
  labels = {
    topic = var.alert_pubsub_topic
  }
}

###############################################################################
# SLOs.
###############################################################################

# Latency SLO: 95% of /v1/score requests < 450ms.
resource "google_monitoring_slo" "fraud_latency" {
  service             = "finguard-${var.environment}-fraud-service"
  slo_id              = "fraud-latency-95p"
  display_name        = "Fraud scoring latency - 95p < 450ms"
  goal                = 0.95
  rolling_period_days = 30

  basic_sli {
    method   = ["REQUEST"]
    location = ["global"]
    version  = ["global"]
    latency {
      threshold = "0.45s"
    }
  }
}

# Availability SLO: 99.9% of /v1/score succeed.
resource "google_monitoring_slo" "fraud_availability" {
  service             = "finguard-${var.environment}-fraud-service"
  slo_id              = "fraud-availability-999"
  display_name        = "Fraud scoring availability - 99.9%"
  goal                = 0.999
  rolling_period_days = 30

  basic_sli {
    method   = ["REQUEST"]
    location = ["global"]
    version  = ["global"]

    availability {
      enabled = true
    }
  }
}

# Document processing latency SLO.
resource "google_monitoring_slo" "doc_latency" {
  service             = "finguard-${var.environment}-document-service"
  slo_id              = "doc-latency-95p"
  display_name        = "Document AI processing - 95p < 5s"
  goal                = 0.95
  rolling_period_days = 30

  basic_sli {
    method   = ["REQUEST"]
    location = ["global"]
    version  = ["global"]
    latency {
      threshold = "5s"
    }
  }
}

###############################################################################
# Alert policies.
###############################################################################

# SLO burn-rate (fast burn).
resource "google_monitoring_alert_policy" "slo_fast_burn" {
  display_name = "Fraud SLO — fast burn (1h)"
  combiner     = "OR"

  conditions {
    display_name = "Fraud availability SLO burn > 14.4x in last 1h"

    condition_threshold {
      filter          = "metric.type=\"monitoring.googleapis.com/services/service/utilization\" AND resource.type=\"service\""
      comparison      = "COMPARISON_GT"
      threshold_value = 14.4
      duration        = "3600s"

      aggregations {
        alignment_period     = "60s"
        per_series_aligner   = "ALIGN_RATE"
        cross_series_reducer = "REDUCE_MEAN"
      }
    }
  }

  notification_channels = [
    google_monitoring_notification_channel.email.id,
    google_monitoring_notification_channel.pubsub.id,
  ]

  alert_strategy {
    auto_close = "86400s" # 1 day
  }
}

# Vertex AI drift detection.
resource "google_monitoring_alert_policy" "vertex_drift" {
  display_name = "Vertex AI fraud model — drift detected"
  combiner     = "OR"

  conditions {
    display_name = "Drift score > 0.3"

    condition_threshold {
      filter          = "metric.type=\"aiplatform.googleapis.com/prediction/feature_drift/attribution_score\" AND resource.type=\"aiplatform.googleapis.com/Endpoint\""
      comparison      = "COMPARISON_GT"
      threshold_value = 0.3
      duration        = "1800s"

      aggregations {
        alignment_period   = "300s"
        per_series_aligner = "ALIGN_MEAN"
      }
    }
  }

  notification_channels = [
    google_monitoring_notification_channel.email.id,
  ]
}

# BigQuery slot saturation.
resource "google_monitoring_alert_policy" "bq_slot_saturation" {
  display_name = "BigQuery slots — high utilization"
  combiner     = "OR"

  conditions {
    display_name = "Slot utilization > 80% for 10 minutes"

    condition_threshold {
      filter          = "metric.type=\"bigquery.googleapis.com/slots/pending\""
      comparison      = "COMPARISON_GT"
      threshold_value = 0.8
      duration        = "600s"

      aggregations {
        alignment_period     = "60s"
        per_series_aligner   = "ALIGN_RATE"
        cross_series_reducer = "REDUCE_MEAN"
      }
    }
  }

  notification_channels = [
    google_monitoring_notification_channel.email.id,
  ]
}
