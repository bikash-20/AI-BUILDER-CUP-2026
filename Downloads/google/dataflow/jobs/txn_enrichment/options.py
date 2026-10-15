"""Pipeline options for the txn_enrichment job."""

from __future__ import annotations

import argparse

import apache_beam as beam
from apache_beam.options.pipeline_options import PipelineOptions, StandardOptions, GoogleCloudOptions, SetupOptions


def build_options(argv: list[str] | None = None) -> PipelineOptions:
    parser = argparse.ArgumentParser()
    parser.add_argument("--input-topic", default="projects/finguard-dev/topics/txn-events")
    parser.add_argument("--output-topic", default="projects/finguard-dev/topics/txn-scored")
    parser.add_argument("--bq-table", default="finguard-dev:fin.transactions")
    parser.add_argument("--runner", default="DirectRunner")
    parser.add_argument("--project", default="finguard-dev")
    parser.add_argument("--region", default="asia-south1")
    parser.add_argument("--temp-location", default="gs://finguard-dataflow-tmp/")
    parser.add_argument("--streaming", action="store_true")
    known, pipeline_args = parser.parse_known_args(argv)

    options = PipelineOptions(pipeline_args, runner=known.runner, save_main_session=True)
    options.view_as(StandardOptions).streaming = bool(known.streaming)
    gcp = options.view_as(GoogleCloudOptions)
    gcp.project = known.project
    gcp.region = known.region
    gcp.job_name = "finguard-txn-enrichment"
    gcp.temp_location = known.temp_location
    options.view_as(SetupOptions).save_main_session = True
    return options, known
