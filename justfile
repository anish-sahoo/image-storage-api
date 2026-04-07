DB_NAME := "image_storage.db"
SCHEMA   := "sql/schema.sql"

# ── Help ──────────────────────────────────────────────────────────────────────

# List available recipes
default:
    @just --list

# ── Setup ─────────────────────────────────────────────────────────────────────

# Full first-time setup: init DB and pull dependencies
setup: db-init deps
    @echo "Setup complete. Run 'just user <username> <password>' to create an admin account."

# Initialise the SQLite database from schema.sql
db-init:
    sqlite3 {{ DB_NAME }} < {{ SCHEMA }}

# Drop and recreate the database
db-reset: db-remove db-init

# Delete the database file
db-remove:
    rm -f {{ DB_NAME }}

# Create an admin user  (usage: just user alice s3cr3t)
user USERNAME PASSWORD:
    go run scripts/create_user.go {{ USERNAME }} {{ PASSWORD }}

# ── Docker ────────────────────────────────────────────────────────────────────

# Start MinIO in the background
up:
    docker compose up -d
    @echo "MinIO API  → http://localhost:9000"
    @echo "MinIO UI   → http://localhost:9001  (minioadmin / minioadmin)"

# Stop MinIO
down:
    docker compose down

# Stop MinIO and wipe its volume
down-clean:
    docker compose down -v

# Tail MinIO logs
logs:
    docker compose logs -f minio

# ── Dev ───────────────────────────────────────────────────────────────────────

# Run the server
run:
    go run .

# Format all Go source files
fmt:
    go fmt ./...

# Tidy go.mod / go.sum
tidy:
    go mod tidy

# Download dependencies
deps:
    go mod download

# Vet the code
vet:
    go vet ./...

# fmt + tidy + vet
lint: fmt tidy vet

# Build the binary
build:
    go build -o bin/image-storage-api .

# Remove build artifacts
clean:
    rm -rf bin/
