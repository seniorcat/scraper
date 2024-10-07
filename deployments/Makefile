# Запуск базы данных с использованием Docker Compose
db_up:
	@echo "Starting database using Docker Compose..."
	docker compose --env-file .env up -d

# Остановка и удаление контейнеров базы данных
db_down:
	@echo "Stopping and removing database containers..."
	docker compose --env-file .env down