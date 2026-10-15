"""Vertex AI Feature Store helpers (point-in-time correct online serving)."""

from __future__ import annotations

import hashlib
from dataclasses import dataclass, field
from typing import Any, Iterator, Protocol

from finguard_common.logging import get_logger

log = get_logger("finguard.feature_store")


@dataclass
class FeatureRow:
    entity_id: str
    features: dict[str, float]


class FeatureStore(Protocol):
    def read(self, entity_ids: list[str], feature_names: list[str]) -> list[FeatureRow]: ...
    def write(self, rows: list[FeatureRow]) -> None: ...


@dataclass
class InMemoryFeatureStore:
    """In-process store useful in tests and local development."""

    _store: dict[str, dict[str, float]] = field(default_factory=dict)

    def read(self, entity_ids: list[str], feature_names: list[str]) -> list[FeatureRow]:
        out: list[FeatureRow] = []
        for eid in entity_ids:
            row = self._store.get(eid, {})
            out.append(FeatureRow(entity_id=eid, features={n: row.get(n, 0.0) for n in feature_names}))
        return out

    def write(self, rows: list[FeatureRow]) -> None:
        for r in rows:
            self._store.setdefault(r.entity_id, {}).update(r.features)

    def iter_all(self) -> Iterator[tuple[str, dict[str, float]]]:
        return iter(self._store.items())


def deterministic_features(seed: str, names: list[str]) -> dict[str, float]:
    """Hash-based deterministic feature values for tests."""
    out = {}
    for n in names:
        h = int(hashlib.sha256(f"{seed}|{n}".encode()).hexdigest()[:8], 16)
        out[n] = (h % 1000) / 1000.0
    return out


def normalize_features(feats: dict[str, float]) -> dict[str, float]:
    """Normalize features to [0, 1] via simple min-max; placeholder."""
    if not feats:
        return feats
    mn, mx = min(feats.values()), max(feats.values())
    if mx == mn:
        return {k: 0.5 for k in feats}
    return {k: (v - mn) / (mx - mn) for k, v in feats.items()}
