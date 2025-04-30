# Booking API

Сервис бронирования рабочих пространств с интеграцией Telegram и Google Calendar.

## Стек технологий

- Go (Golang)
- PostgreSQL
- Gin (HTTP API)
- Docker, docker-compose
- Telegram Bot API
- Google Calendar API

## Diagram
- \diagram

## Возможности

- Регистрация и аутентификация пользователей (JWT)
- CRUD для рабочих пространств
- Бронирование рабочего места с проверкой времени, конфликтов и расчетом цены
- Отмена бронирования (штраф за позднюю отмену)
- Уведомления в Telegram и интеграция с Google Calendar(также в телеграмме через инлайн кнопки можно создавать, отменять и посмотреть все бронирования)
- Подробное логирование

## Быстрый старт

### 1. Клонируйте репозиторий

```sh
git clone <???>
```

### 2. Настройте переменные окружения

Создайте файл `.env` по примеру:

```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=booking
DB_SSLMODE=disable
TELEGRAM_BOT_TOKEN=...
GOOGLE_CALENDAR_CREDENTIALS=...
GOOGLE_CALENDAR_ID=...
JWT_SECRET=your_jwt_secret
```

### 3. Миграции 
```
psql -h localhost -U postgres -d booking_service -f internal/infrastructure/repository/migrations/01_init_tables.sql
```

### 4. Запуск через Docker Compose

```sh
docker-compose up --build
```

API будет доступен на `http://localhost:8080`.


## Тесты

Для запуска unit-тестов:

```sh
go test -v ./internal/usecase/...
```

## Документация API

- Swagger: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html) (если настроено)
- Примеры запросов — см. коллекцию Postman в папке `docs/`

## Контакты

- Автор: Anuar Oralov
- Telegram: [@anuaroralov](https://t.me/anuaroralov)
