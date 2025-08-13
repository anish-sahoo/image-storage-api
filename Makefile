DB_NAME=image_storage.db
SCHEMA=sql/schema.sql

.PHONY: db-init db-remove create-data-storage data-remove run

db-init:
	sqlite3 $(DB_NAME) < $(SCHEMA)

db-remove:
	rm image_storage.db

create-data-storage:
	mkdir -p data/

data-remove:
	rm -rf data/

run:
	go run .