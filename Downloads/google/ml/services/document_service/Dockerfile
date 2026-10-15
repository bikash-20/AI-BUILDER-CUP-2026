# syntax=docker/dockerfile:1.7
FROM python:3.11-slim

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PYTHONFAULTHANDLER=1 \
    PIP_NO_CACHE_DIR=1

WORKDIR /app

# OS packages
COPY packages.txt* ./
RUN if [ -f packages.txt ]; then \
    apt-get update && apt-get install -y --no-install-recommends $(cat packages.txt) && rm -rf /var/lib/apt/lists/*; \
    fi

# Install Python deps
COPY requirements.txt .
RUN pip install -r requirements.txt

# Copy libraries and service
COPY ../../libs /app/libs
COPY . /app

ENV PYTHONPATH=/app:/app/libs

EXPOSE 8080
ENTRYPOINT ["uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8080"]
