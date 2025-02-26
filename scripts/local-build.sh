#!/bin/bash

# Script to build and run the Webline Backend locally

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Webline Backend Local Development ===${NC}"
echo ""

# Check for Docker and Docker Compose
command -v docker >/dev/null 2>&1 || { echo -e "${RED}Docker is required but not installed. Aborting.${NC}" >&2; exit 1; }

# Determine Docker Compose command
if command -v docker-compose >/dev/null 2>&1; then
  DOCKER_COMPOSE="docker-compose"
elif docker compose version >/dev/null 2>&1; then
  DOCKER_COMPOSE="docker compose"
else
  echo -e "${RED}Neither docker-compose nor docker compose is available. Aborting.${NC}" >&2
  exit 1
fi

# Check if .env file exists, create if not
if [ ! -f ".env" ]; then
  echo -e "${YELLOW}No .env file found. Creating a basic one for local development.${NC}"
  cat > .env << EOF
# Database Configuration
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=webline
DB_HOST=db
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=webline
PORT=8080

# Environment
ENV=development

# REDIS
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_POOL_SIZE=20
REDIS_MIN_IDLE_CONNS=5
REDIS_TTL=15m
REDIS_RATE_LIMIT=50

# Set other variables as needed for local development
# ...
EOF
  echo -e "${GREEN}.env file created with default values.${NC}"
  echo -e "${YELLOW}Please update it with your specific configuration if needed.${NC}"
  echo ""
fi

# Modify docker-compose.yml to use local build for backend
echo -e "${BLUE}Building and starting containers...${NC}"
sed -i.bak 's|image: \${DOCKER_HUB_USERNAME}/webline-backend:\${IMAGE_TAG:-latest}|build: .|g' docker-compose.yml || {
  # If sed fails (macOS compatibility)
  sed -i '' 's|image: \${DOCKER_HUB_USERNAME}/webline-backend:\${IMAGE_TAG:-latest}|build: .|g' docker-compose.yml || {
    echo -e "${YELLOW}Could not automatically modify docker-compose.yml. Please manually change:${NC}"
    echo -e "${YELLOW}  image: \${DOCKER_HUB_USERNAME}/webline-backend:\${IMAGE_TAG:-latest}${NC}"
    echo -e "${YELLOW}to:${NC}"
    echo -e "${YELLOW}  build: .${NC}"
    echo ""
    read -p "Press Enter to continue after making the change or Ctrl+C to abort..."
  }
}

# Build and start containers
echo -e "${BLUE}Starting the application...${NC}"
$DOCKER_COMPOSE up --build -d

# Check container status
echo -e "${BLUE}Checking container status...${NC}"
$DOCKER_COMPOSE ps

echo ""
echo -e "${GREEN}Application should be running now!${NC}"
echo -e "${BLUE}Access the API at:${NC} http://localhost:8080"
echo -e "${BLUE}To view logs:${NC} $DOCKER_COMPOSE logs -f"
echo -e "${BLUE}To stop the application:${NC} $DOCKER_COMPOSE down"
echo ""
