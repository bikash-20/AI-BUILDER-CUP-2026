# organization

Creates the FinGuard AI folder hierarchy and applies a baseline of
organization policies. The `org_id` variable is used as the binding target
for `google_organization_policy` resources; it must be passed in from the
calling environment (Terraform cannot discover organization IDs implicitly).

This module is intentionally small — most org-level changes require manual
console interaction with an Org Admin account. See the top-level
`terraform/bootstrap/README.md` for what cannot be terraformed.
