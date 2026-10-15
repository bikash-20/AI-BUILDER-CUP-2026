"""Minimal FastAPI to test if Python 3.14 + uvicorn works."""
import uvicorn
from fastapi import FastAPI

app = FastAPI()


@app.get("/healthz")
def h():
    return {"status": "ok"}


if __name__ == "__main__":
    uvicorn.run(app, host="127.0.0.1", port=8081, log_level="info")
