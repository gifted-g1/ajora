
#!/bin/bash
set -e

echo "Building all Ajora services..."

# Build Go services
echo "Building Go services..."
for service in services/*-service; do
  if [ -d "$service" ]; then
    echo "  Building $service..."
    cd "$service"
    go build -o bin/service .
    cd - > /dev/null
  fi
done

# Build Rust services
echo "Building Rust services..."
cd services/wallet-signer
cargo build --release
cd - > /dev/null

echo "✅ All services built!"

