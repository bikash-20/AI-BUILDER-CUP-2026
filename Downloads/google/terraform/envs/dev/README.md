# dev environment

This is the **development** environment. It's the cheapest, least
constrained configuration. In particular:

- `min_instances` for Cloud Run services are 0 (scales to zero).
- AlloyDB is **not** provisioned (we use SQLite in tests for dev).
- All resources are in a single project under the `Dev` folder.
- The state bucket prefix is `envs/dev`.

To apply:

```bash
cp backend.tf.example backend.tf  # fill in bucket name
cp terraform.tfvars.example terraform.tfvars  # fill in values
terraform init
terraform plan -out tfplan
terraform apply tfplan
```
