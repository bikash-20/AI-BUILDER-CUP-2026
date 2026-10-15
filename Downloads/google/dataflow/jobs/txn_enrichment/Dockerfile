# syntax=docker/dockerfile:1.7
FROM apache/beam_python3.11_sdk:2.55.0
WORKDIR /app
COPY requirements.txt .
RUN pip install -r requirements.txt
COPY . /app
ENTRYPOINT ["python", "/app/pipeline.py"]
