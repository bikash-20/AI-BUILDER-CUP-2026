"""Autoencoder training for anomaly detection on transactions.

Used in the ensemble alongside XGBoost. Reconstruction error > threshold ⇒
anomalous transaction.
"""

from __future__ import annotations

import argparse
import json
import os
import sys

from finguard_common.logging import configure_logging, get_logger
from finguard_evaluation.metrics import binary_metrics

configure_logging()
log = get_logger("finguard.training.autoencoder")


def main() -> int:
    p = argparse.ArgumentParser()
    p.add_argument("--model-dir", default=os.getenv("AIP_MODEL_DIR", "/tmp/ae-model"))
    p.add_argument("--epochs", type=int, default=5)
    p.add_argument("--mock", action="store_true", default=os.getenv("MOCK_MODE", "true") == "true")
    args = p.parse_args()

    metrics = _train_mock(epochs=args.epochs) if args.mock else _train_real(args)
    os.makedirs(args.model_dir, exist_ok=True)
    with open(os.path.join(args.model_dir, "metrics.json"), "w") as f:
        json.dump(metrics, f, indent=2)
    log.info("autoencoder_train_done", metrics=metrics)
    return 0


def _train_mock(epochs: int) -> dict[str, float]:
    try:
        import numpy as np
        import tensorflow as tf
    except Exception as e:  # pragma: no cover
        return {"error": str(e)}

    rng = np.random.default_rng(0)
    X = rng.normal(size=(300, 8))
    model = tf.keras.Sequential(
        [
            tf.keras.layers.Dense(4, activation="relu", input_shape=(8,)),
            tf.keras.layers.Dense(8, activation="linear"),
        ]
    )
    model.compile(optimizer="adam", loss="mse")
    model.fit(X, X, epochs=epochs, verbose=0)
    recon = model.predict(X, verbose=0)
    mse = ((X - recon) ** 2).mean(axis=1)
    # Synthetic labels: high MSE = anomaly
    y_true = (mse > mse.mean()).astype(int).tolist()
    y_score = (mse / mse.max()).tolist()
    m = binary_metrics(y_true, y_score)
    m["epochs"] = epochs
    return m


def _train_real(args) -> dict[str, float]:
    # Real implementation would pull from BigQuery and persist to Vertex Model Registry.
    return {"status": "real training not implemented in scaffold"}


if __name__ == "__main__":
    sys.exit(main())
