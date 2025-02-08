# =============== MIGRATIONS =========================
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


# =======================================================
swag:
	swag init -g ./cmd/main.go -o ./swagger

test:
	go test ./tests/...

test-v:
	go test ./tests/... -v