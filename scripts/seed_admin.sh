#!/usr/bin/env bash
set -euo pipefail
# API sẽ tự seed admin nếu không tồn tại, dựa vào ENV:
# API_ADMIN_EMAIL / API_ADMIN_PASSWORD
echo "Admin will be seeded by API on first start via ENV."
echo "Make sure to set API_ADMIN_EMAIL and API_ADMIN_PASSWORD in .env file"
