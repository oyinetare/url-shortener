#!/bin/sh

# Load environment variables from .env file if it exists
if [ -f .env ]; then
    echo "Loading environment variables from .env file..."
    export $(grep -v '^#' .env | xargs)
fi

# Set defaults if not provided
DATABASE_HOST=${DATABASE_HOST:-127.0.0.1}
DATABASE_PORT=${DATABASE_PORT:-3306}
DATABASE_NAME=${DATABASE_NAME:-urls}
DATABASE_USER=${DATABASE_USER:-url_shorten_service}
DATABASE_PASSWORD=${DATABASE_PASSWORD:-123}
MYSQL_ROOT_PASSWORD=${MYSQL_ROOT_PASSWORD:-123}

# Clean up any existing container
docker stop urls_db 2>/dev/null
docker rm urls_db 2>/dev/null

# Run the MySQL container with platform specification
echo "Starting DB..."
echo "Database: $DATABASE_NAME"
echo "User: $DATABASE_USER"

docker run --name urls_db -d \
  -e MYSQL_ROOT_PASSWORD="$MYSQL_ROOT_PASSWORD" \
  -e MYSQL_DATABASE="$DATABASE_NAME" \
  -e MYSQL_USER="$DATABASE_USER" \
  -e MYSQL_PASSWORD="$DATABASE_PASSWORD" \
  -p "$DATABASE_PORT:3306" \
  mysql:8.0

# Wait for the database service to start up
echo "Waiting for DB to start up..."
until docker exec urls_db mysqladmin --silent -u"$DATABASE_USER" -p"$DATABASE_PASSWORD" ping; do
  echo "Waiting for database connection..."
  sleep 2
done

# Give MySQL a bit more time to fully initialize
sleep 5

# Run the setup script if it exists
if [ -f setup.sql ]; then
    echo "Setting up initial data..."
    docker exec -i urls_db mysql -u"$DATABASE_USER" -p"$DATABASE_PASSWORD" "$DATABASE_NAME" < setup.sql
fi

echo "Database setup complete!"
echo "Connection details:"
echo "  Host: localhost"
echo "  Port: $DATABASE_PORT"
echo "  Database: $DATABASE_NAME"
echo "  User: $DATABASE_USER"
