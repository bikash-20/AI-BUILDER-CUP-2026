# network

Provisions a **hub-and-spoke** VPC topology:

- A single VPC per environment (`<env>-vpc`) with **private** subnets across
  two regions (primary + DR).
- **Cloud NAT** for egress (no public IPs on workloads).
- **Private Service Access** for connecting to Google-managed services
  (BigQuery, Cloud SQL, AlloyDB) over private IP.
- **Private Google Access** enabled on all subnets so VMs without external IPs
  can still reach `*.googleapis.com`.
- A handful of **VPC firewall rules** for ingress/egress; everything else
  is denied by default (default VPC has implicit deny on ingress).

A separate **Cloud Router + NAT** is configured per region for high availability.
