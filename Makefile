.PHONY: up down logs build seed gen-secret

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f --tail=200

build:
	# ví dụ build UI rồi copy vào volume (tùy pipeline của bạn)
	echo "Build in CI and publish to ghcr.io"

seed:
	./scripts/seed_admin.sh

gen-secret:
	./scripts/gen_jwt_secret.sh
