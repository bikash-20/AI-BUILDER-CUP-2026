###############################################################################
# VPC
###############################################################################

resource "google_compute_network" "vpc" {
  name                    = "${var.environment}-finguard-vpc"
  auto_create_subnetworks = false
  routing_mode            = "REGIONAL"

  # Soft-delete protection is enabled at the project level; this VPC is
  # the foundation for everything else, so we lock it down.
  delete_default_routes_on_create = false
}

###############################################################################
# Subnets — one primary, one DR, both /20 with private Google access.
###############################################################################

resource "google_compute_subnetwork" "primary" {
  name          = "${var.environment}-primary"
  ip_cidr_range = "10.10.0.0/20"
  region        = var.primary_region
  network       = google_compute_network.vpc.id

  private_ip_google_access = true

  secondary_ip_range = [
    {
      range_name    = "pods"
      ip_cidr_range = "10.20.0.0/16"
    },
    {
      range_name    = "services"
      ip_cidr_range = "10.30.0.0/16"
    },
  ]
}

resource "google_compute_subnetwork" "dr" {
  name          = "${var.environment}-dr"
  ip_cidr_range = "10.11.0.0/20"
  region        = var.dr_region
  network       = google_compute_network.vpc.id

  private_ip_google_access = true

  secondary_ip_range = [
    {
      range_name    = "pods"
      ip_cidr_range = "10.21.0.0/16"
    },
    {
      range_name    = "services"
      ip_cidr_range = "10.31.0.0/16"
    },
  ]
}

###############################################################################
# Cloud Router + NAT for egress.
###############################################################################

resource "google_compute_router" "primary" {
  name    = "${var.environment}-primary-router"
  region  = google_compute_subnetwork.primary.region
  network = google_compute_network.vpc.id
}

resource "google_compute_router_nat" "primary" {
  name                               = "${var.environment}-primary-nat"
  router                             = google_compute_router.primary.name
  region                             = google_compute_router.primary.region
  nat_ip_allocate_option             = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"

  log_config {
    enable = true
    filter = "ERRORS_ONLY"
  }
}

resource "google_compute_router" "dr" {
  name    = "${var.environment}-dr-router"
  region  = google_compute_subnetwork.dr.region
  network = google_compute_network.vpc.id
}

resource "google_compute_router_nat" "dr" {
  name                               = "${var.environment}-dr-nat"
  router                             = google_compute_router.dr.name
  region                             = google_compute_router.dr.region
  nat_ip_allocate_option             = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"

  log_config {
    enable = true
    filter = "ERRORS_ONLY"
  }
}

###############################################################################
# Private Service Access for Google-managed services (BigQuery, Cloud SQL,
# AlloyDB) over private IP. Requires a /16 reserved range.
###############################################################################

resource "google_compute_global_address" "private_services" {
  name          = "${var.environment}-private-services"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = google_compute_network.vpc.id
}

resource "google_service_networking_connection" "private_services" {
  network                 = google_compute_network.vpc.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.private_services.name]
}

###############################################################################
# Firewall rules — explicit allow only; everything else is implicit deny.
###############################################################################

# IAP for SSH / RDP — restricts to IAP's published range.
resource "google_compute_firewall" "allow_iap" {
  name    = "${var.environment}-allow-iap-ssh"
  network = google_compute_network.vpc.id

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }

  # IAP's TCP forwarding netblock.
  source_ranges = ["35.235.240.0/20"]
  target_tags   = ["ssh-iap-enabled"]
}

# Internal traffic between subnets in the VPC.
resource "google_compute_firewall" "allow_internal" {
  name    = "${var.environment}-allow-internal"
  network = google_compute_network.vpc.id

  allow {
    protocol = "tcp"
    ports    = ["0-65535"]
  }
  allow {
    protocol = "udp"
    ports    = ["0-65535"]
  }
  allow {
    protocol = "icmp"
  }

  source_ranges = [var.primary_cidr, var.dr_cidr]
}
