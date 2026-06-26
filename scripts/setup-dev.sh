
#!/bin/bash
set -e

echo "Setting up Ajora development environment..."

# Check prerequisites
command -v go >/dev/null 2>&1 || { echo "Go is required"; exit 1; }
command -v rustc >/dev/null 2>&1 || { echo "Rust is required"; exit 1; }
command -v docker >/dev/null 2>&1 || { echo "Docker is required"; exit 1; }

# Install dependencies
echo "Installing Go dependencies..."
for service in services/*; do
  if [ -f "$service/go.mod" ]; then
    echo "  $service"
    cd "$service"
    go mod download
    cd - > /dev/null
  fi
done

echo "Installing Rust dependencies..."
for service in services/wallet-signer; do
  if [ -f "$service/Cargo.toml" ]; then
    echo "  $service"
    cd "$service"
    cargo fetch
    cd - > /dev/null
  fi
done

# Start services
echo "Starting PostgreSQL and Redis..."
docker-compose up -d postgres redis

echo "Waiting for databases to be ready..."
sleep 10

echo "Running migrations..."
docker exec -i ajora-postgres psql -U ajora_admin -d ajora < migrations/001_initial_schema.sql

echo "✅ Setup complete!"

