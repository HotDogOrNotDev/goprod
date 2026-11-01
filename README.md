# Практическое задание: Безопасный сервис аутентификации

## 🎯 Цель задания

Разработать безопасный REST API сервис с функциями регистрации и аутентификации пользователей на Go.

## 📋 Что нужно реализовать

### Обязательный функционал:
- ✅ **Регистрация пользователя** с хешированием пароля (bcrypt)
- ✅ **Вход в систему** с выдачей JWT токена
- ✅ **Защищенный эндпоинт** для получения профиля (требует JWT)
- ✅ **Защита от SQL-инъекций** (параметризованные запросы)

### API эндпоинты:
| Метод | Путь | Описание | Требует токен |
|-------|------|----------|--------------|
| POST | `/register` | Регистрация пользователя | Нет |
| POST | `/login` | Вход в систему | Нет |
| GET | `/profile` | Получить профиль | **Да** |
| GET | `/health` | Проверка состояния | Нет |

## 🏗️ Структура проекта

```
secure-service/
├── main.go              # Главный файл с запуском сервера
├── handlers.go          # HTTP обработчики
├── models.go            # Структуры данных
├── database.go          # Работа с БД
├── auth.go              # JWT и bcrypt
├── middleware.go        # Проверка токена
├── docker-compose.yml   # PostgreSQL в Docker
├── init.sql             # Схема БД
├── .env                 # Конфигурация (создать из .env.example)
├── go.mod               # Зависимости
└── README.md           # Этот файл
```

## 🚀 Быстрый старт

### 1. Настройка окружения

```bash
# Создайте .env файл из примера
cp .env.example .env

# ВАЖНО: Измените JWT_SECRET в .env на свой ключ (минимум 32 символа)
nano .env
```

### 2. Запуск базы данных

```bash
# Запустите PostgreSQL в Docker
docker-compose up -d

# Проверьте, что БД запустилась
docker-compose ps
```

### 3. Установка зависимостей

```bash
# Скачайте Go модули
go mod download
```

### 4. Что нужно реализовать

Все файлы с пометкой TODO содержат заготовки функций, которые нужно завершить:

#### 📄 `database.go` - Работа с базой данных
- [x] `CreateUser()` - создание пользователя
- [x] `GetUserByEmail()` - поиск по email
- [x] `GetUserByID()` - поиск по ID
- [x] `UserExistsByEmail()` - проверка существования

#### 🔐 `auth.go` - Аутентификация и безопасность
- [x] `HashPassword()` - хеширование паролей bcrypt
- [x] `CheckPassword()` - проверка паролей
- [x] `GenerateToken()` - создание JWT токенов
- [x] `ValidateToken()` - проверка JWT токенов

#### 🛡️ `middleware.go` - Защита эндпоинтов
- [x] `AuthMiddleware()` - проверка токенов

#### 🌐 `handlers.go` - HTTP обработчики
- [x] `RegisterHandler()` - регистрация
- [x] `LoginHandler()` - авторизация
- [x] `ProfileHandler()` - профиль пользователя

## 📝 Пошаговое руководство

### Шаг 1: Реализуйте функции безопасности (`auth.go`)

```go
// Импортируйте необходимые пакеты
import (
    "golang.org/x/crypto/bcrypt"
    "github.com/golang-jwt/jwt/v5"
)

// Реализуйте HashPassword
func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}
```

### Шаг 2: Реализуйте работу с БД (`database.go`)

```go
// ВАЖНО: Используйте параметризованные запросы!
func CreateUser(email, username, passwordHash string) (*User, error) {
    query := `INSERT INTO users (email, username, password_hash) VALUES ($1, $2, $3) RETURNING id, created_at`
    // Реализуйте...
}
```

### Шаг 3: Реализуйте middleware (`middleware.go`)

```go
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 1. Получите токен из заголовка Authorization
        // 2. Проверьте формат "Bearer <token>"
        // 3. Валидируйте токен
        // 4. Добавьте данные в контекст
        // 5. Передайте управление дальше
    }
}
```

### Шаг 4: Реализуйте обработчики (`handlers.go`)

Каждый обработчик содержит детальные комментарии с пошаговыми инструкциями.

### Шаг 5: Запустите и протестируйте

```bash
# Запустите сервер
go run *.go

# В другом терминале тестируйте API
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","username":"testuser","password":"SecurePass123"}'
```

## 🧪 Тестирование API

### 1. Проверка здоровья сервиса
```bash
curl http://localhost:8080/health
```

### 2. Регистрация пользователя
```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "username": "testuser",
    "password": "SecurePass123"
  }'
```

### 3. Вход в систему
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123"
  }'
```

### 4. Получение профиля (с токеном)
```bash
# Замените YOUR_JWT_TOKEN на токен из ответа /login
curl http://localhost:8080/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 🔒 Требования безопасности

### ✅ Обязательные требования:

1. **Пароли хешируются bcrypt**
   ```go
   // ❌ НЕПРАВИЛЬНО
   user.Password = password

   // ✅ ПРАВИЛЬНО
   hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
   ```

2. **SQL запросы параметризованы**
   ```go
   // ❌ ОПАСНО - SQL инъекции!
   query := fmt.Sprintf("SELECT * FROM users WHERE email = '%s'", email)

   // ✅ БЕЗОПАСНО
   query := "SELECT * FROM users WHERE email = $1"
   db.QueryRow(query, email)
   ```

3. **JWT токены проверяются**
   ```go
   // ❌ БЕЗ ПРОВЕРКИ
   func ProfileHandler(w http.ResponseWriter, r *http.Request) {
       // Сразу возвращаем данные
   }

   // ✅ С ПРОВЕРКОЙ
   http.HandleFunc("/profile", AuthMiddleware(ProfileHandler))
   ```

## 🐛 Частые ошибки

### 1. Пароли в открытом виде
```sql
-- ❌ ПЛОХО: пароль не захеширован
SELECT password_hash FROM users; -- "123456"

-- ✅ ХОРОШО: bcrypt хеш
-- "$2a$10$N9qo8uLOickgx2ZMRZoMye..."
```

### 2. SQL инъекции
```go
// ❌ УЯЗВИМО
query := "SELECT * FROM users WHERE email = '" + email + "'"

// ✅ ЗАЩИЩЕНО
query := "SELECT * FROM users WHERE email = $1"
db.QueryRow(query, email)
```

### 3. JWT не проверяется
```go
// ❌ ОПАСНО
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
    // Нет проверки токена!
}

// ✅ БЕЗОПАСНО
http.HandleFunc("/profile", AuthMiddleware(ProfileHandler))
```

## ✅ Чек-лист перед сдачей

- [x] PostgreSQL запускается через `docker-compose up`
- [x] Приложение подключается к БД и не падает
- [x] Регистрация создает пользователя в БД
- [x] Пароли хранятся как bcrypt хеш, НЕ в открытом виде
- [x] Вход возвращает валидный JWT токен
- [x] Токен можно декодировать на https://jwt.io
- [x] Эндпоинт `/profile` требует токен (без токена → 401)
- [x] Эндпоинт `/profile` работает с правильным токеном
- [x] **ВСЕ** SQL запросы используют параметры `$1, $2...`
- [x] В коде НЕТ `fmt.Sprintf` для построения SQL

## 🔍 Проверка безопасности

### Проверьте хеширование паролей:
```bash
# Подключитесь к БД
docker exec -it secure_service_db psql -U postgres -d secure_service

# Проверьте хеши паролей
SELECT email, password_hash FROM users;

# Хеш должен начинаться с $2a$ или $2b$
\q
```

### Проверьте JWT токен:
1. Скопируйте токен из ответа `/login`
2. Вставьте на https://jwt.io
3. Убедитесь, что содержит `user_id`, `email`, `username`

## 🆘 Получение помощи

### Если что-то не работает:

1. **БД не запускается**
   ```bash
   docker-compose down
   docker-compose up -d
   docker-compose logs postgres
   ```

2. **Ошибки компиляции**
   ```bash
   go mod tidy
   go mod download
   ```

3. **Сервер не запускается**
   - Проверьте .env файл
   - Убедитесь, что JWT_SECRET длиннее 32 символов
   - Проверьте, что PostgreSQL запущен

4. **Тесты API не проходят**
   - Проверьте логи сервера
   - Убедитесь, что все TODO функции реализованы
   - Проверьте правильность JSON в curl запросах

## 🎯 Критерии оценки

### "Зачёт" - все требования выполнены:
- ✅ Регистрация и авторизация работают
- ✅ Пароли хешируются bcrypt
- ✅ JWT токены используются правильно
- ✅ SQL запросы параметризованы
- ✅ Защищенные эндпоинты требуют токен
- ✅ Код компилируется и запускается

### "На доработку":
- ❌ Пароли в открытом виде
- ❌ SQL инъекции возможны
- ❌ JWT не проверяются
- ❌ Код не компилируется
