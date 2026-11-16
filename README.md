# GitHub Repository Indexer

Сервис на Go для загрузки и хранения информации о GitHub-репозиториях.

Приложение состоит из двух сервисов:

- **Gateway** — принимает HTTP-запросы и возвращает JSON.
- **Collector** — загружает данные из GitHub API и сохраняет их в PostgreSQL.

Gateway обращается к Collector через gRPC. Массовая загрузка выполняется через worker pool с ограничением параллелизма, таймаутами и отменой через context.

**Стек:** Go, PostgreSQL, Docker Compose, gRPC, Protocol Buffers, REST API, Swagger.

## Возможности

- Получение информации о репозитории: название, описание, звёзды, форки и дата создания.
- Хранение данных в PostgreSQL: повторный GET возвращает сохранённую запись без обращения к GitHub.
- Массовая загрузка и обновление репозиториев с отдельным результатом для каждого элемента.

Код сервисов находится в `internal/collector` и `internal/gateway`, точки входа — в `cmd`.

## Запуск через Docker Compose

```bash
docker compose up --build -d
```

API доступно по адресу `http://localhost:8080`.

Миграции goose применяются автоматически при запуске Collector. PostgreSQL хранит данные в постоянном томе.

Остановка:

```bash
docker compose down
```

Данные сохраняются после остановки и повторного запуска.

## Запуск без Docker

Для Collector нужна работающая PostgreSQL:

```bash
export DATABASE_URL='postgres://indexer:indexer@localhost:5432/indexer?sslmode=disable'
go run ./cmd/collector
```

В другом терминале:

```bash
go run ./cmd/gateway
```

## Примеры запросов

Получение одного репозитория:

```bash
curl http://localhost:8080/repositories/golang/go
```

Массовая загрузка:

```bash
curl -X POST http://localhost:8080/repositories \
  -H 'Content-Type: application/json' \
  -d '{"repositories":[{"owner":"golang","repo":"go"},{"owner":"docker","repo":"cli"}]}'
```

В ответе `results` каждый элемент содержит `owner`, `repo` и либо `data`, либо `error`. Ошибка одного репозитория не останавливает остальные. Повторный POST обновляет сохранённые данные.

## Настройки

| Переменная | По умолчанию | Назначение |
| --- | --- | --- |
| `DATABASE_URL` | задаётся вручную; в Compose настроена | Подключение Collector к PostgreSQL |
| `COLLECTOR_ADDR` | `localhost:50051` | Адрес Collector; в Compose `collector:50051` |
| `WORKER_COUNT` | `4` | Число workers на массовую загрузку |
| `GITHUB_TIMEOUT` | `10s` | Таймаут запроса к GitHub |
| `BATCH_TIMEOUT` | `60s` | Общий таймаут массовой загрузки |
| `REQUEST_TIMEOUT` | `70s` | Таймаут запроса в Gateway |

Для Compose значения можно указать в `.env`. Внешний порт Gateway задаётся через `GATEWAY_PORT` (по умолчанию `8080`), пароль PostgreSQL — через `POSTGRES_PASSWORD` (по умолчанию `indexer`).

## Swagger

Документация API: `http://localhost:8080/swagger/index.html`.

## Лицензия
MIT