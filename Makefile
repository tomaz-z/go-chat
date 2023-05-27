.PHONY: start stop logs

.SILENT: start stop logs

start:
	docker compose up -d

stop:
	docker compose down

logs:
	docker compose logs
