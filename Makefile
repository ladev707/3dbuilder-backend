# Learning note: .PHONY marks targets that are command names, not files.
# Without it, a file named "routes" in this folder would stop `make routes`
# from running.
.PHONY: debug routes migrate-up migrate-down migrate-status migrate-create

debug:
	dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient --log ./cmd/api

routes:
	go run ./cmd/console routes:show

migrate-up:
	go run ./cmd/migrator up

migrate-down:
	go run ./cmd/migrator down

migrate-status:
	go run ./cmd/migrator status

# Usage: make migrate-create name=add_products
migrate-create:
	@test -n "$(name)" || (echo "usage: make migrate-create name=create_users_table" && exit 1)
	goose -dir migrations -s create $(name) sql
