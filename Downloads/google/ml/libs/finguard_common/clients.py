"""Lazy-loaded GCP client factories.

Each factory respects `MOCK_MODE` and returns a stub when in mock mode,
allowing local execution without credentials.
"""

from __future__ import annotations

from typing import Any

from finguard_common.config import is_mock_mode, project_id


def _maybe_client(import_path: str, klass: str, mock_factory: Any) -> Any:
    """Try to import a real client; fall back to mock factory."""
    if is_mock_mode():
        return mock_factory()
    try:
        module_path, _, class_name = import_path.rpartition(".")
        module = __import__(module_path, fromlist=[class_name])
        return getattr(module, klass)()
    except Exception:
        return mock_factory()


# ----- Mocks ----------------------------------------------------------------

class MockBigQueryClient:
    def query(self, *args: Any, **kwargs: Any) -> "MockQuery":  # noqa: D401
        return MockQuery()


class MockQuery:
    def result(self) -> list[dict[str, Any]]:
        return []


class MockDocumentAIClient:
    def process_document(self, *args: Any, **kwargs: Any) -> "MockDoc":
        return MockDoc()


class MockDoc:
    document = None


class MockDLPClient:
    def deidentify_content(self, *args: Any, **kwargs: Any) -> "MockDLPRsp":  # noqa: D401
        return MockDLPRsp()


class MockDLPRsp:
    item = []


class MockPubSubPublisher:
    def publish(self, *args: Any, **kwargs: Any) -> "str":
        return "mock-message-id"


class MockSecretManagerClient:
    def access_secret_version(self, *args: Any, **kwargs: Any) -> "MockSecret":  # noqa: D401
        return MockSecret()


class MockSecret:
    payload = type("P", (), {"data": b'{"mock": true}'})()


class MockStorageClient:
    def bucket(self, *args: Any, **kwargs: Any) -> "MockBucket":
        return MockBucket()


class MockBucket:
    def blob(self, *args: Any, **kwargs: Any) -> "MockBlob":
        return MockBlob()


class MockBlob:
    def upload_from_string(self, *args: Any, **kwargs: Any) -> None:
        return None


class MockVertexClient:
    class Predict:  # noqa: D401
        def predict(self, *args: Any, **kwargs: Any) -> list[dict[str, Any]]:
            return []


# ----- Public factories -----------------------------------------------------

def bigquery_client() -> Any:
    return _maybe_client(
        "google.cloud.bigquery",
        "Client",
        lambda: MockBigQueryClient(),
    )


def documentai_client() -> Any:
    return _maybe_client(
        "google.cloud.documentai",
        "DocumentProcessorServiceClient",
        lambda: MockDocumentAIClient(),
    )


def dlp_client() -> Any:
    return _maybe_client(
        "google.cloud.dlp_v2",
        "DlpServiceClient",
        lambda: MockDLPClient(),
    )


def pubsub_publisher() -> Any:
    return _maybe_client(
        "google.cloud.pubsub_v1",
        "PublisherClient",
        lambda: MockPubSubPublisher(),
    )


def secret_manager_client() -> Any:
    return _maybe_client(
        "google.cloud.secretmanager",
        "SecretManagerServiceClient",
        lambda: MockSecretManagerClient(),
    )


def storage_client() -> Any:
    return _maybe_client(
        "google.cloud.storage",
        "Client",
        lambda: MockStorageClient(),
    )


def vertex_client() -> Any:
    from finguard_common.config import region as _region

    if is_mock_mode():
        return MockVertexClient()
    try:
        from google.cloud import aiplatform  # type: ignore

        aiplatform.init(project=project_id(), location=_region())
        return aiplatform
    except Exception:
        return MockVertexClient()
