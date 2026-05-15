import json
from unittest.mock import patch, MagicMock

import pytest

from app import app


@pytest.fixture
def client():
    app.config["TESTING"] = True
    with app.test_client() as c:
        yield c


def test_health(client):
    resp = client.get("/health")
    assert resp.status_code == 200
    data = resp.get_json()
    assert data["status"] == "healthy"
    assert data["service"] == "gateway"
    assert "uptime_seconds" in data


def test_list_services(client):
    resp = client.get("/api/services")
    assert resp.status_code == 200
    data = resp.get_json()
    assert len(data["services"]) == 3
    names = [s["name"] for s in data["services"]]
    assert "gateway" in names
    assert "collector" in names
    assert "dashboard" in names


@patch("app.requests.get")
def test_aggregate_status_all_healthy(mock_get, client):
    mock_resp = MagicMock()
    mock_resp.json.return_value = {"status": "healthy"}
    mock_get.return_value = mock_resp
    resp = client.get("/api/status")
    assert resp.status_code == 200
    data = resp.get_json()
    assert data["gateway"]["status"] == "healthy"
    assert data["collector"]["status"] == "healthy"
    assert data["dashboard"]["status"] == "healthy"


@patch("app.requests.get")
def test_aggregate_status_service_down(mock_get, client):
    import requests as req
    mock_get.side_effect = req.RequestException("connection refused")
    resp = client.get("/api/status")
    assert resp.status_code == 200
    data = resp.get_json()
    assert data["collector"]["status"] == "unreachable"
    assert data["dashboard"]["status"] == "unreachable"
    assert data["gateway"]["status"] == "healthy"


def test_forward_metrics_no_body(client):
    resp = client.post("/api/metrics", content_type="application/json")
    assert resp.status_code == 400


@patch("app.requests.post")
def test_forward_metrics_success(mock_post, client):
    mock_resp = MagicMock()
    mock_resp.json.return_value = {"received": True}
    mock_resp.status_code = 200
    mock_post.return_value = mock_resp
    resp = client.post("/api/metrics", json={"cpu": 42.5})
    assert resp.status_code == 200
    data = resp.get_json()
    assert data["received"] is True


@patch("app.requests.post")
def test_forward_metrics_collector_down(mock_post, client):
    import requests as req
    mock_post.side_effect = req.RequestException("timeout")
    resp = client.post("/api/metrics", json={"cpu": 42.5})
    assert resp.status_code == 502
    data = resp.get_json()
    assert data["error"] == "collector unreachable"
