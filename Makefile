include .env

# ================================================ LOCAL =====================================
install:
	docker compose up --build -d
	docker compose down;

start:
	docker compose up -d ast-db

stop:
	docker compose down;



# migrations
mc: # $(name)
	docker compose run --rm ast-goose create $(name) sql
	sudo chmod -R 777 migrations/

m:
	docker compose run --rm ast-goose up

m-one:
	docker compose run --rm ast-goose up-by-one

md: # down one last migration
	docker compose run --rm ast-goose down

md-to: # $(timestamp) - откат конкретной миграции. пример: make md-to timestamp=20170506082527
	docker compose run --rm ast-goose down-to $(timestamp)

mockery:
	docker compose run --rm ast-mockery

# CLI
cli-clean-db:
	go run cmd/cli/main.go clean-db;



# ================================================ PRODUCTION ===========================================

prod-start:
	docker compose -f docker-compose.prod.yaml up --build -d

prod-stop:
	docker compose -f docker-compose.prod.yaml down

prod-bash:
	docker exec -it ast-app sh

deploy:
	git pull;
	make prod-start;
	sleep 5;
	make prod-m;
	@if [ $$(df -BG --output=avail / | tail -n 1 | tr -d 'G') -lt 3 ]; then \
		echo "Мало места (<3GB), чистим Docker..."; \
		yes | docker system prune --volumes -f; \
	else \
		echo "Места достаточно, пропускаем очистку."; \
	fi


# migrations
prod-m:
	docker compose -f docker-compose.prod.yaml run --rm ast-goose up

prod-m-one:
	docker compose -f docker-compose.prod.yaml run --rm ast-goose up-by-one

prod-md: # down one last migration
	docker compose -f docker-compose.prod.yaml run --rm ast-goose down

prod-md-to: # $(timestamp) - откат конкретной миграции. пример: make md-to timestamp=20170506082527
	docker compose -f docker-compose.prod.yaml run --rm ast-goose down-to $(timestamp)


#CLI
cli-clean-db-p:
	docker exec ast-app ./cliApp clean-db;

# =============== BACKUP/RESTORE =========================

backup-db:
	docker exec ast-db pg_dump -U $(DB_USERNAME) -d $(DB_DATABASE) | gzip > $(DB_LOCAL_BACKUP_PATH)/$(shell date +%Y-%m-%d_%H-%M-%S).sql.gz
	chown -R $(DB_LOCAL_BACKUP_OWNER):$(DB_LOCAL_BACKUP_OWNER) $(DB_LOCAL_BACKUP_PATH);
	echo "Database backup created successfully"

db-remove-old-backups: # Удаляет бэкапы БД, которые были созданы более 5 дней назад
	find $(DB_LOCAL_BACKUP_PATH) -type f -mtime +5 -exec rm -rf {} +

restore-db: # with param file=path/to/backup/dump.sql
	gunzip -c $(file) | docker exec -i ast-db psql -U $(DB_USERNAME) -d $(DB_DATABASE)
	echo "Database restored successfully"

# =======================================================
swag:
	swag init -g ./cmd/main.go -o ./swagger

test:
	go test ./tests/...

test-v:
	go test ./tests/... -v