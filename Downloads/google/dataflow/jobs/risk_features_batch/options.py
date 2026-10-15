"""Pipeline options for the risk_features_batch job."""

from __future__ import annotations

import argparse

from apache_beam.options.pipeline_options import GoogleCloudOptions, PipelineOptions, SetupOptions, StandardOptions


def build_options(argv: list[str] | None = None) -> tuple[PipelineOptions, argparse.Namespace]:
    parser = argparse.ArgumentParser()
    parser.add_argument("--bq-source", default="finguard-dev:fin.transactions")
    parser.add_argument("--bq-dest", default="finguard-dev:fin.customer_features_mv")
    parser.add_argument("--runner", default="DirectRunner")
    parser.add_argument("--project", default="finguard-dev")
    parser.add_argument("--region", default="asia-south1")
    parser.add_argument("--temp-location", default="gs://finguard-dataflow-tmp/")
    known, pipeline_args = parser.parse_known_args(argv)

    options = PipelineOptions(pipeline_args, runner=known.runner, save_main_session=True)
    options.view_as(StandardOptions).streaming = False
    gcp = options.view_as(GoogleCloudOptions)
    gcp.project = known.project
    gcp.region = known.region
    gcp.job_name = "finguard-risk-features"
    gcp.temp_location = known.temp_location
    options.view_as(SetupOptions).save_main_session = True
    return options, known
