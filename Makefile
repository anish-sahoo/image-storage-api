DB_NAME=image_storage.db
SCHEMA=sql/schema.sql

.PHONY: db-init
db-init:
	sqlite3 $(DB_NAME) < $(SCHEMA)

db-remove:
	rm -f $(DB_NAME)
