#!/usr/bin/env bash
set -euo pipefail
echo "Generated JWT Secret (32 bytes hex):"
openssl rand -hex 32
echo ""
echo "Copy this value to API_JWT_SECRET in your .env file"
