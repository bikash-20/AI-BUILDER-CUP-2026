"""Standard classification metrics.

These functions are pure-Python (no numpy/sklearn dependency at import-time)
so they're cheap to import in any environment.
"""

from __future__ import annotations

import math
from collections.abc import Iterable, Sequence


def binary_metrics(
    y_true: Sequence[int],
    y_score: Sequence[float],
    threshold: float = 0.5,
) -> dict[str, float]:
    """Compute precision, recall, F1, AUC for binary classification.

    AUC is computed via the rank-sum method (Mann-Whitney U).
    """
    if len(y_true) != len(y_score):
        raise ValueError("y_true and y_score must have the same length")
    if not y_true:
        return {"precision": 0.0, "recall": 0.0, "f1": 0.0, "auc_roc": 0.0}

    tp = fp = fn = tn = 0
    for yt, ys in zip(y_true, y_score):
        pred = 1 if ys >= threshold else 0
        if pred == 1 and yt == 1:
            tp += 1
        elif pred == 1 and yt == 0:
            fp += 1
        elif pred == 0 and yt == 1:
            fn += 1
        else:
            tn += 1

    precision = tp / (tp + fp) if tp + fp else 0.0
    recall = tp / (tp + fn) if tp + fn else 0.0
    f1 = (2 * precision * recall / (precision + recall)) if (precision + recall) else 0.0

    auc = _auc(y_true, y_score)
    return {"precision": precision, "recall": recall, "f1": f1, "auc_roc": auc}


def _auc(y_true: Sequence[int], y_score: Sequence[float]) -> float:
    positives = [s for yt, s in zip(y_true, y_score) if yt == 1]
    negatives = [s for yt, s in zip(y_true, y_score) if yt == 0]
    if not positives or not negatives:
        return 0.5
    wins = 0.0
    for p in positives:
        for n in negatives:
            if p > n:
                wins += 1
            elif p == n:
                wins += 0.5
    return wins / (len(positives) * len(negatives))


def expected_calibration_error(
    y_true: Sequence[int],
    y_score: Sequence[float],
    n_bins: int = 10,
) -> float:
    """Compute ECE (Expected Calibration Error)."""
    n = len(y_true)
    if n == 0:
        return 0.0
    bin_size = 1.0 / n_bins
    ece = 0.0
    for b in range(n_bins):
        lo, hi = b * bin_size, (b + 1) * bin_size
        idx = [i for i, s in enumerate(y_score) if lo <= s < hi or (b == n_bins - 1 and s == hi)]
        if not idx:
            continue
        acc = sum(y_true[i] for i in idx) / len(idx)
        conf = sum(y_score[i] for i in idx) / len(idx)
        ece += abs(acc - conf) * (len(idx) / n)
    return ece


def population_stability_index(expected: Sequence[float], actual: Sequence[float], n_bins: int = 10) -> float:
    """PSI: drift measure between two distributions.

    Returns a number; PSI > 0.2 is typically considered significant drift.
    """
    if not expected or not actual:
        return 0.0

    bins = _bin_edges(expected, n_bins)
    e_counts = _bin_counts(expected, bins)
    a_counts = _bin_counts(actual, bins)

    e_total = sum(e_counts) or 1
    a_total = sum(a_counts) or 1

    psi = 0.0
    for e, a in zip(e_counts, a_counts):
        e_share = (e + 1e-6) / e_total
        a_share = (a + 1e-6) / a_total
        psi += (a_share - e_share) * math.log(a_share / e_share)
    return psi


def _bin_edges(values: Iterable[float], n_bins: int) -> list[float]:
    sorted_v = sorted(values)
    n = len(sorted_v)
    edges = []
    for i in range(n_bins + 1):
        idx = min(n - 1, int(i * (n - 1) / n_bins))
        edges.append(sorted_v[idx])
    return sorted(set(edges))


def _bin_counts(values: Sequence[float], edges: list[float]) -> list[int]:
    counts = [0] * (len(edges) - 1)
    for v in values:
        for i in range(len(edges) - 1):
            if edges[i] <= v < edges[i + 1]:
                counts[i] += 1
                break
        else:
            if v >= edges[-1]:
                counts[-1] += 1
    return counts
