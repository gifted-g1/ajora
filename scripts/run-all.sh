
#!/bin/bash
set -e

echo "Starting all Ajora services..."

# Start infrastructure
docker-compose up -d postgres redis

# Start services
for service in services/*-service; do
  if [ -f "$service/bin/service" ]; then
    echo "Starting $service..."
    cd "$service"
    ./bin/service &
    cd - > /dev/null
  fi
done

# Start wallet signer
echo "Starting wallet signer..."
cd services/wallet-signer
cargo run --release &
cd - > /dev/null

echo "✅ All services started!"
echo "Services running on ports:"
echo "  Auth Service: 8081"
echo "  User Service: 8082"
echo "  Pool Service: 8083"
echo "  Wallet Signer: 8088"

