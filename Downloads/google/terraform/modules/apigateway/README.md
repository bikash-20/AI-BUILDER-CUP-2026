# apigateway

Provisions an API Gateway in front of the FinGuard AI Cloud Run services.
The gateway:

- Exposes a single HTTPS endpoint with a managed SSL certificate.
- Routes traffic to the appropriate Cloud Run service based on the path.
- Performs JWT verification (Identity Platform) at the edge.
- Adds request size limits and CORS preflight handling.

The OpenAPI spec in `openapi.yaml` is the source of truth for the
contract — Terraform imports it as a `google_api_gateway_api_config`
resource. CI regenerates the spec from `docs/api-reference.md` (see
`scripts/` for tooling).
