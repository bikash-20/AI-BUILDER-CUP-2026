"""Vertex AI Search RAG retrieval.

Builds a context string by retrieving the top-k documents from a Vertex AI
Search datastore.
"""

from __future__ import annotations

from typing import Any, Protocol

from finguard_common.config import get_settings, is_mock_mode
from finguard_common.logging import get_logger

log = get_logger("finguard.gemini.rag")


class VertexSearchRetriever(Protocol):
    def retrieve(self, query: str, top_k: int = 5) -> list[dict[str, Any]]: ...


class MockRetriever:
    """Returns canned snippets for tests/local development."""

    def retrieve(self, query: str, top_k: int = 5) -> list[dict[str, Any]]:
        text = query.lower()
        if "decline" in text or "declined" in text:
            return [
                {
                    "title": "Card decline decisioning",
                    "uri": "policy://risk/geo-velocity-v2",
                    "content": "Cards are declined when geo-velocity exceeds 1000 km/h or device risk > 0.8.",
                }
            ]
        if "kyc" in text:
            return [
                {
                    "title": "KYC policy v3",
                    "uri": "policy://kyc/v3",
                    "content": "All KYC docs require PAN + address proof + selfie matching.",
                }
            ]
        return []


class VertexSearchRetrieverImpl:
    """Calls Vertex AI Search via google-cloud-discoveryengine."""

    def __init__(self, datastore: str | None = None) -> None:
        s = get_settings()
        self.datastore = datastore or s.vertex_search_datastore
        self._client: Any = None
        if not is_mock_mode() and self.datastore:
            try:
                from google.cloud import discoveryengine  # type: ignore

                self._client = discoveryengine.SearchServiceClient()
            except Exception as e:  # pragma: no cover
                log.warning("vertex_search_init_failed", error=str(e))

    def retrieve(self, query: str, top_k: int = 5) -> list[dict[str, Any]]:
        if self._client is None:
            return MockRetriever().retrieve(query, top_k)
        try:
            s = get_settings()
            serving = f"projects/{s.project_id}/locations/global/collections/default_collection/dataStores/{self.datastore}/servingConfigs/default_config"
            request = {
                "servingConfig": serving,
                "query": query,
                "pageSize": top_k,
            }
            response = self._client.search(request=request)
            return [
                {"title": r.document.title, "uri": r.document.uri, "content": str(r.document.derived_struct_data)}
                for r in response
            ]
        except Exception as e:
            log.warning("vertex_search_failed", error=str(e))
            return MockRetriever().retrieve(query, top_k)


def build_rag_context(query: str, top_k: int = 5) -> tuple[str, list[str]]:
    """Return (context, citations)."""
    retriever: VertexSearchRetriever = VertexSearchRetrieverImpl()
    docs = retriever.retrieve(query, top_k=top_k)
    if not docs:
        return "", []
    ctx_parts = [f"[{i+1}] {d['title']}\n{d['content']}" for i, d in enumerate(docs)]
    citations = [d["uri"] for d in docs]
    return "\n\n".join(ctx_parts), citations
