
.PHONY: start

NOW = $(shell date -u '+%Y%m%d%I%M%S')
PWD=$(shell pwd)
UID=$(shell id -u)


app/init: ## Init app (create config files, init git submodules, etc.)
app/init: ## Init app (create config files, init git submodules, etc.)
	@if [ ! -f .env ]; then cp .env.dist .env; fi
	@. ./.env && if test "$$ENV" = "dev" && test ! -f docker-compose.override.yml ; then \
	    cp docker-compose.override.dev.yml docker-compose.override.yml; \
    fi
	@git rev-parse --verify HEAD | head -c8 > version


app/restart: ## Restart application services
app/restart: app/stop app/start

app/start: ## Start application services for current environment
app/start: app/init docker/network/create
	docker-compose up -d
	$(MAKE) docker/migrate/up

app/stop: ## Stop application services
	@docker-compose down --remove-orphans

all: start

build:
	go build -o ./bin/skeleton ./cmd/api/main.go

clean:
	rm -fr ./bin/skeleton

start: clean
	go run cmd/api/main.go  -c ./configs/api.yaml -swagger ./api/openapi_spec

docker/network/create:
	docker network create skeleton || true

docker/build/skeleton:
	docker-compose build skeleton

docker/build/migration:
	docker-compose build migration

docker/migrate/up: docker/network/create
	docker-compose run --rm migration sql-migrate up -env=db

docker/migrate/down: docker/network/create
	docker-compose run --rm migration sql-migrate down -env=db

docker/migrate/new:
	docker-compose run --rm migration bash -c "sql-migrate new -env=db ${name} && chown -R ${UID}:${UID} /opt/migrations"

docker/build:
	docker-compose build

