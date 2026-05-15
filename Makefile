.PHONY: up down test test-gateway test-collector test-dashboard lint build

up:
	docker compose up --build -d

down:
	docker compose down

build:
	docker compose build

test: test-gateway test-collector test-dashboard

test-gateway:
	cd gateway && pip install -q -r requirements.txt && pytest -v

test-collector:
	cd collector && go test -v ./...

test-dashboard:
	cd dashboard && npm install --silent && npm test

lint: lint-gateway lint-collector lint-dashboard

lint-gateway:
	cd gateway && flake8 --max-line-length=120 app.py

lint-collector:
	cd collector && go vet ./...

lint-dashboard:
	cd dashboard && npm install --silent && npx eslint src/

logs:
	docker compose logs -f

status:
	@curl -s http://localhost:8000/health | python3 -m json.tool
	@curl -s http://localhost:8001/health | python3 -m json.tool
	@curl -s http://localhost:8002/health | python3 -m json.tool

clean:
	docker compose down -v --rmi local
