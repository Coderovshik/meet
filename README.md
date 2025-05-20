# Meet

Meet - это веб-приложение для проведения видеозвонков.

[Ссылка для преподавателя с демо-версией приложения](https://amogus.root-hub.com)

## 🚀 Технологии

### Frontend
- React.js
- Vite

### Backend
- Go (Golang)
- Redis
- Docker
- WebRTC
- REST API

## 🏃‍♂️ Быстрый старт

### Предварительные требования
- Docker и Docker Compose
- Node.js (для разработки frontend)
- Go 1.24+ (для разработки backend)

### Запуск с помощью Docker

1. Клонируйте репозиторий:
```bash
git clone https://github.com/your-username/meet.git
cd meet
```

2. Запустите приложение:
```bash
cd backend
docker-compose up -d
```

Приложение будет доступно по адресу: http://localhost:8080

### Разработка

#### Frontend
```bash
cd frontend
npm install
npm run dev
```

#### Backend
```bash
cd backend
go mod download
go run cmd/meet/main.go
```

## 📁 Структура проекта

```
meet/
├── frontend/           # React приложение
│   ├── src/           # Исходный код
│   ├── public/        # Статические файлы
│   └── dist/          # Собранное приложение
│
└── backend/           # Go сервер
    ├── cmd/          # Точка входа приложения
    ├── internal/     # Внутренние пакеты
    ├── web/          # Статические файлы
    ├── Dockerfile    # Конфигурация Docker
    └── docker-compose.yml # Конфигурация Docker Compose
```

## 📝 Лицензия

MIT 