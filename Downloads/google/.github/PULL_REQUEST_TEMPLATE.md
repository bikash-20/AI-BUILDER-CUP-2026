## Summary

<!-- 1-3 sentences. What does this PR do and why? -->

## Related

<!-- Link to issue(s), ADRs, design docs, etc. -->

## Test plan

- [ ] Unit tests added/updated
- [ ] Integration test added (if applicable)
- [ ] Manual test steps documented below
- [ ] Terraform plan reviewed
- [ ] Cloud Deploy pipeline previewed (if applicable)

## Manual test steps

<!-- If the change isn't fully covered by automated tests, list
     the steps a reviewer can take to verify. -->

## Risk & rollback

- **Risk level:** (Low / Medium / High)
- **Rollback plan:** (Describe how to roll back. If the change is a
  Cloud Run revision, mention `gcloud run services update-traffic`.)

## Checklist

- [ ] I have read [CONTRIBUTING.md](../blob/main/docs/CONTRIBUTING.md)
- [ ] I have added tests for new behavior
- [ ] I have updated the OpenAPI spec if I changed the API contract
- [ ] I have updated docs/ if I changed architecture or operations
