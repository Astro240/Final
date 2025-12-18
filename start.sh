#!/bin/bash

# Initialize database directories if they don't exist
mkdir -p backend/database
mkdir -p backend/store_images/banners
mkdir -p backend/store_images/logos
mkdir -p backend/store_images/products
mkdir -p backend/avatars

# Start Docker containers
docker-compose up -d

echo "Astropify is starting..."
echo ""
echo "Access on your network:"
echo "  http://astropify.com"
echo "  http://www.astropify.com"
echo "  http://api.astropify.com"
echo ""
echo "View logs with: docker-compose logs -f backend"
echo "Stop with: docker-compose down"
