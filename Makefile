MIGRATE ?= migrate
DB_DSN ?= postgres://kalistheniks:kalistheniks@localhost:5432/kalistheniks?sslmode=disable
MIGRATIONS_DIR := migrations

.PHONY: migrate migrate-up migrate-down migrate-force migrate-create

migrate-up:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DB_DSN)" up

migrate-down:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DB_DSN)" down

migrate-force:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DB_DSN)" force ${version}

migrate-create:
	$(MIGRATE) create -ext sql -dir $(MIGRATIONS_DIR) ${name}
