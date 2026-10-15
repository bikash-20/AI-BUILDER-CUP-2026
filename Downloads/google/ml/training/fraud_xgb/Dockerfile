# syntax=docker/dockerfile:1.7
# Vertex AI Custom Training container for XGBoost fraud model.
FROM python:3.11-slim
ENV PYTHONDONTWRITEBYTECODE=1 PYTHONUNBUFFERED=1 PIP_NO_CACHE_DIR=1
WORKDIR /app
COPY requirements.txt .
RUN pip install -r requirements.txt
COPY ../../libs /app/libs
COPY . /app
ENV PYTHONPATH=/app:/app/libs
ENTRYPOINT ["python", "train.py"]
