###############################################################################
# Cluster.
###############################################################################

resource "google_alloydb_cluster" "platform" {
  cluster_id   = "finguard-${var.environment}"
  location     = var.region
  network      = var.network_id

  # CCMEK encryption.
  encryption_config {
    kms_key_name = var.alloydb_kms_key_id
  }

  # Automated backups + PITR.
  automated_backup_policy {
    enabled  = true
    location = var.region
    weekly_schedule {
      days_of_week = ["MONDAY", "WEDNESDAY", "FRIDAY"]
      start_times {
        hours   = 2
        minutes = 0
        seconds = 0
        nanos   = 0
      }
    }
  }

  # Continuous backup (PITR).
  continuous_backup_config {
    enabled              = true
    recovery_window_days = 14
    encryption_config {
      kms_key_name = var.alloydb_kms_key_id
    }
  }

  labels = {
    env        = var.environment
    managed_by = "terraform"
  }
}

###############################################################################
# Primary instance.
###############################################################################

resource "google_alloydb_instance" "primary" {
  cluster       = google_alloydb_cluster.platform.name
  instance_id   = "finguard-${var.environment}-primary"
  instance_type = "PRIMARY"

  machine_config {
    cpu_count = var.environment == "prod" ? 8 : 2
  }

  database_flags = {
    "alloydb.enable_pgaudit"       = "on"
    "alloydb.lock_wait_timeout"    = "20s"
    "alloydb.log_min_duration_statement" = "1000"
  }

  labels = {
    env = var.environment
  }
}

###############################################################################
# Read pool (prod only — staging skips this for cost).
###############################################################################

resource "google_alloydb_instance" "read_pool" {
  count = var.environment == "prod" ? 1 : 0

  cluster       = google_alloydb_cluster.platform.name
  instance_id   = "finguard-${var.environment}-read-pool"
  instance_type = "READ_POOL"

  machine_config {
    cpu_count = 4
  }

  read_pool_config {
    node_count = 2
  }

  labels = {
    env = var.environment
  }
}
