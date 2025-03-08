include .env

# =============== PRODUCTION =========================
prod-start:
	docker compose -f docker-compose.prod.yaml up --build -d

prod-stop:
	docker compose -f docker-compose.prod.yaml down

prod-bash:
	docker exec -it ast-app sh

deploy:
	git pull;
	make prod-start;
	make prod-m;

# =============== MIGRATIONS =========================
# prod
prod-m:
	docker exec -it ast-app goose up;

prod-m-one:
	docker exec -it ast-app goose up-by-one;

prod-md: # down one last migration
	docker exec -it ast-app goose down;

prod-md-to: # $(timestamp) - откат конкретной миграции. пример: make md-to timestamp=20170506082527
	docker exec -it ast-app goose down-to $(timestamp);


# local
mc: # $(name)
	goose create $(name) sql

m:
	goose up

m-one:
	goose up-by-one

md: # down one last migration
	goose down

md-to: # $(timestamp) - откат конкретной миграции. пример: make md-to timestamp=20170506082527
	goose down-to $(timestamp)

# =============== BACKUP/RESTORE =========================

backup-db:
	docker exec ast-db pg_dump -U $(DB_USERNAME) -d $(DB_DATABASE) > $(DB_LOCAL_BACKUP_PATH)/$(shell date +%Y-%m-%d_%H-%M-%S).sql
	chown -R $(DB_LOCAL_BACKUP_OWNER):$(DB_LOCAL_BACKUP_OWNER) $(DB_LOCAL_BACKUP_PATH);
	echo "Database backup created successfully"

db-remove-old-backups: # Удаляет бэкапы БД, которые были созданы более 5 дней назад
	find $(DB_LOCAL_BACKUP_PATH) -type f -mtime +5 -exec rm -rf {} +

restore-db: # with param file=path/to/backup/dump.sql
	docker exec -i ast-db psql -U $(DB_USERNAME) -d $(DB_DATABASE) < $(file)
	echo "Database restored successfully"

# =======================================================
swag:
	swag init -g ./cmd/main.go -o ./swagger

mockery:
	mockery

test:
	go test ./tests/...

test-v:
	go test ./tests/... -v