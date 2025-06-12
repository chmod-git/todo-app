# Todo-App

**Todo-App** is a backend project built in **Go** that combines a **REST API** (using Gin) and a **Telegram bot** (using `go-telegram-bot-api`) to help users manage to-do lists and tasks.

---

## Tech Stack

* **Go**, **Gin** — RESTful API
* **PostgreSQL** — Primary database
* **Redis** — Caching
* **Telegram Bot API** — Task control via chat
* **JWT** — Authentication
* **Docker**, **.env** — Configuration & deployment

---

## Project Structure

```
cmd/            # Entry point
configs/        # App configs
pkg/            # Handlers, services, repositories
tg_bot/         # Telegram bot logic
schema/         # DB init SQL
tests/          # Unit, integration, functional tests
```

---

## Core Features

* **Users**: Register, login, edit, delete account (via API or bot)
* **Todo Lists**: Create, update, view, delete
* **Todo Items**: Add, edit, mark as done/undone
* **Telegram bot**: Chat interface for managing tasks

---

## Telegram Commands

```
/start      - Launch bot
/account    - Edit or delete your profile
/lists      - Manage to-do lists and items
```

---

## Setup

Create a `.env` file:

```
DB_HOST=localhost
DB_PORT=5432
...
BOT_TOKEN=your_telegram_bot_token
```

Run the app:

```bash
go run cmd/main.go         # Start API
go run tg_bot/tg_bot.go    # Start Telegram bot
```

