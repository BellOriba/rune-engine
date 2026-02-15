.PHONY: up down restart logs stats redis-cli clean

up:
	docker compose up -d --build

down:
	docker compose down

restart:
	docker compose restart rune-engine-proxy

logs:
	docker compose logs -f

stats:
	docker stats

redis-cli:
	docker exec -it rune-engine-redis redis-cli -a ${REDIS_PASSWORD}

clean:
	docker system prune -f
	docker volume prune -f
