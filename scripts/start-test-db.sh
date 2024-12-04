#!/bin/bash

# コンテナ名
CONTAINER_NAME="column-exporter-test-db"

# 既存のコンテナを停止・削除
docker stop $CONTAINER_NAME >/dev/null 2>&1
docker rm $CONTAINER_NAME >/dev/null 2>&1

# MySQLコンテナを起動
docker run --name $CONTAINER_NAME \
  -e MYSQL_ROOT_PASSWORD=password \
  -e MYSQL_DATABASE=test_export \
  -v "$(pwd)/testdata/init.sql:/docker-entrypoint-initdb.d/init.sql" \
  -p 3306:3306 \
  -d mysql:8.0 \
  --character-set-server=utf8mb4 \
  --collation-server=utf8mb4_unicode_ci

echo "Starting MySQL container..."
echo "Waiting for database initialization..."
sleep 20

echo "Database is ready!"
echo "You can now test the exporter with this command:"
echo
echo "Export profile_image column:"
echo "./column-exporter export \\"
echo "  --driver mysql \\"
echo "  --host localhost \\"
echo "  --port 3306 \\"
echo "  --user root \\"
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