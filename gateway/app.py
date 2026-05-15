import logging
import os
import time

from flask import Flask, jsonify, request
import requests

app = Flask(__name__)

LOG_LEVEL = os.environ.get("GATEWAY_LOG_LEVEL", "INFO").upper()
logging.basicConfig(
    level=getattr(logging, LOG_LEVEL, logging.INFO),
    format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
)
logger = logging.getLogger("gateway")

COLLECTOR_URL = os.environ.get("COLLECTOR_URL", "http://localhost:8001")
DASHBOARD_URL = os.environ.get("DASHBOARD_URL", "http://localhost:8002")

START_TIME = time.time()


@app.route("/health")
def health():
    return jsonify({"status": "healthy", "service": "gateway", "uptime_seconds": round(time.time() - START_TIME, 2)})


@app.route("/api/services", methods=["GET"])
def list_services():
    logger.info("Listing registered services")
    services = [
        {"name": "gateway", "url": "http://gateway:8000", "type": "api-gateway"},
        {"name": "collector", "url": COLLECTOR_URL, "type": "metrics-collector"},
        {"name": "dashboard", "url": DASHBOARD_URL, "type": "dashboard-api"},
    ]
    return jsonify({"services": services})


@app.route("/api/status", methods=["GET"])
def aggregate_status():
    logger.info("Aggregating status from all services")
    results = {}
    for name, url in [("collector", COLLECTOR_URL), ("dashboard", DASHBOARD_URL)]:
        try:
            resp = requests.get(f"{url}/health", timeout=5)
            results[name] = resp.json()
        except requests.RequestException as e:
            logger.warning("Failed to reach %s: %s", name, e)
            results[name] = {"status": "unreachable", "error": str(e)}
    results["gateway"] = {"status": "healthy", "uptime_seconds": round(time.time() - START_TIME, 2)}
    return jsonify(results)


@app.route("/api/metrics", methods=["POST"])
def forward_metrics():
    data = request.get_json()
    if not data:
        logger.error("No JSON body provided to /api/metrics")
        return jsonify({"error": "request body must be JSON"}), 400
    logger.info("Forwarding metrics to collector")
    try:
        resp = requests.post(f"{COLLECTOR_URL}/metrics", json=data, timeout=5)
        return jsonify(resp.json()), resp.status_code
    except requests.RequestException as e:
        logger.error("Failed to forward metrics: %s", e)
        return jsonify({"error": "collector unreachable", "details": str(e)}), 502


def create_app():
    return app


if __name__ == "__main__":
    port = int(os.environ.get("GATEWAY_PORT", 8000))
    logger.info("Starting gateway on port %d", port)
    app.run(host="0.0.0.0", port=port)
