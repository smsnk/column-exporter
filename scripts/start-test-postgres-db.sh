#!/bin/bash

# コンテナ名
CONTAINER_NAME="column-exporter-test-postgres-db"

# 既存のコンテナを停止・削除
docker stop $CONTAINER_NAME >/dev/null 2>&1
docker rm $CONTAINER_NAME >/dev/null 2>&1

# PostgreSQLコンテナを起動
docker run --name $CONTAINER_NAME \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=test_export \
  -v "$(pwd)/testdata/init_postgres.sql:/docker-entrypoint-initdb.d/init.sql" \
  -p 5432:5432 \
  -d postgres:14

echo "Starting PostgreSQL container..."
echo "Waiting for database initialization..."
sleep 20

echo "Database is ready!"
echo "You can now test the exporter with this command:"
echo
echo "Export profile_image column:"
echo "./column-exporter export \\"
echo "  --driver postgres \\"
echo "  --host localhost \\"
echo "  --port 5432 \\"
echo "  --user postgres \\"
echo "  --password password \\"
echo "  --database test_export \\"
echo "  --table users \\"
echo "  --column profile_image \\"
echo "  --name-column username \\"
echo "  --output ./image_output \\"
echo "  --ext .bin"
echo
echo "To stop the database:"
echo "docker stop $CONTAINER_NAME" 