.PHONY: start stop logs build/% run/%

start:
	docker compose up --detach

stop:
	docker compose down

logs:
	docker compose logs

build/%:
	go build -o /bin/$* /app/src/cmd/$*

run/%:
	/bin/$*

test:
	$(MAKE) -C src/cmd/client start
