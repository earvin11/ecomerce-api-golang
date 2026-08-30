# Ecommerce API

API REST de e-commerce en **Go**, con autenticación JWT y arquitectura limpia. Pensada como backend para tiendas online: gestión de categorías, productos, roles y usuarios, con login/refresh de sesiones.

## Características

- REST API bajo `/api/v1`, router `net/http` estándar (patrones de método + path params).
- Autenticación JWT (HS256): login con `username` + `password` (bcrypt) y refresh de sesión.
- Tokens de acceso (`JWT_ACCESS_TTL`, default 30m) y refresh (`JWT_REFRESH_TTL`, default 720h).
- Middleware de autorización `Authorization: Bearer <token>` en todas las rutas salvo los endpoints públicos.
- Bootstrap opcional de un usuario `ADMIN` inicial al arrancar.
- Paginación normalizada en listados (`meta`).
- Respuestas JSON estandarizadas: `{message, data, error, meta?}`.
- Validación de entrada (JSON estricto, sin campos desconocidos, validaciones por recurso).
- Schema aplicado automáticamente al arrancar (idempotente) y roles sembrados (`ADMIN`, `CUSTOMER`, `EDITOR`).

## Stack

| Capa | Tecnología |
|------|------------|
| Lenguaje | Go 1.26.5 |
| Router | `net/http` (stdlib, Go 1.22+ method patterns) |
| Base de datos | PostgreSQL 16 (pgx/v5 `pgxpool`) |
| Auth | `golang-jwt/jwt/v5` (HS256) + `golang.org/x/crypto` (bcrypt) |
| Infra | Docker / Docker Compose |

## Arquitectura

El proyecto sigue arquitectura limpia: las dependencias apuntan hacia adentro y la inyección se hace únicamente en `cmd/api/main.go`.

```
cmd/api/main.go          # wiring + arranque HTTP + graceful shutdown
internal/config/         # configuración por variables de entorno
internal/domain/         # entidades, DTOs y errores (sin dependencias externas)
internal/auth/           # servicio JWT (firma y parseo de tokens)
internal/repository/     # interfaces + implementaciones pgx + schema.sql
internal/use_cases/      # lógica de negocio (consumen interfaces)
internal/handlers/       # handlers HTTP, router y middleware de auth
```

Flujo de imports: `handlers` → `use_cases` → `repository` → `domain`.

## Requisitos

- **Go 1.26+** para correr localmente.
- **Docker + Docker Compose** (recomendado para la base de datos y/o el stack completo).

## Configuración

Toda la configuración se lee de variables de entorno (ver `.env.example`).

| Variable | Default | Descripción |
|----------|---------|-------------|
| `SERVER_PORT` | `8080` | Puerto del servidor HTTP |
| `READ_TIMEOUT` / `WRITE_TIMEOUT` / `IDLE_TIMEOUT` | `10s` / `10s` / `60s` | Timeouts del servidor |
| `DB_HOST` / `DB_PORT` | `localhost` / `5432` | Host y puerto de PostgreSQL |
| `DB_USER` / `DB_PASSWORD` | `postgres` / `postgres` | Credenciales de la base |
| `DB_NAME` | `ecommerce` | Nombre de la base |
| `DB_SSL_MODE` | `disable` | Modo SSL de conexión |
| `DB_MAX_CONNS` | `10` | Conexiones máximas del pool |
| `JWT_SECRET` | **obligatorio** | Secreto para firmar tokens (usa un valor largo y aleatorio) |
| `JWT_ACCESS_TTL` | `30m` | Vigencia del token de acceso |
| `JWT_REFRESH_TTL` | `720h` | Vigencia del token de refresh |
| `SEED_ADMIN_USERNAME` / `SEED_ADMIN_EMAIL` / `SEED_ADMIN_PASSWORD` | vacío | Opcional: crean el primer admin al arrancar. **Las 3 deben definirse juntas** |

## Cómo levantar la API

### Opción 1 — Stack completo con Docker Compose (recomendado)

1. Crea tu archivo de entorno a partir del ejemplo y define un `JWT_SECRET`:

   ```bash
   cp .env.example .env
   # edita .env y cambia JWT_SECRET por un valor largo y aleatorio
   # (opcional) define SEED_ADMIN_USERNAME, SEED_ADMIN_EMAIL y SEED_ADMIN_PASSWORD
   ```

2. Construye y levanta la API + PostgreSQL:

   ```bash
   docker compose up --build
   ```

3. La API queda disponible en `http://localhost:8080`.

> **Nota:** el schema se aplica con `CREATE TABLE IF NOT EXISTS`. Si cambia el schema en `internal/repository/schema.sql` y ya existe un volumen previo, es necesario recrearlo para que se apliquen los cambios: `docker compose down -v` y volver a `docker compose up --build` (esto **borra los datos**).

### Opción 2 — Desarrollo local (API fuera de Docker, Postgres en Docker)

1. Levanta solo la base de datos:

   ```bash
   docker compose up -d db
   ```

2. Define las variables de entorno necesarias (PowerShell p.ej.):

   ```powershell
   $env:JWT_SECRET = "tu-secreto-largo-y-aleatorio"
   # opcional, junto:
   $env:SEED_ADMIN_USERNAME = "admin"
   $env:SEED_ADMIN_EMAIL    = "admin@example.com"
   $env:SEED_ADMIN_PASSWORD = "admin123456"
   ```

   (Si tus credenciales de Postgres no coinciden con los defaults, define también `DB_USER`, `DB_PASSWORD`, `DB_NAME`, etc.)

3. Arranca la API:

   ```bash
   go run ./cmd/api
   ```

En ambas opciones el arranque aplica el schema, siembra los roles y (si está configurado) crea el usuario admin inicial.

## Comandos útiles

| Comando | Descripción |
|---------|-------------|
| `go build ./...` | Compila el proyecto |
| `go vet ./...` | Analiza el código |
| `go test ./...` | Ejecuta los tests |
| `gofmt -l .` | Verifica formato (debe quedar vacío) |

## API

### Formato de respuesta

Todas las respuestas son JSON con la forma `{message, data, error}`.

- **Éxito (un recurso):** `{"message": "...", "data": {...}, "error": null}`
- **Éxito (listado):** `{"message": "...", "data": [...], "meta": {...}, "error": null}`
- **Error:** `{"message": "...", "data": null, "error": "detalle"}`

`meta` (paginación): `{total, page, per_page, last_page, remaining}`. Parámetros de listado: `?page=1&page_size=10` (default `page=1`, `page_size=10`, máx `100`).

Códigos de estado: `201` creado · `200` get/update · `204` delete (sin cuerpo) · `400` bad request · `401` no autorizado · `404` not found · `500` internal error.

### Autenticación

- Endpoints públicos: `GET /health`, `POST /api/v1/auth/login`, `POST /api/v1/auth/refresh`.
- **Todos** los demás endpoints requieren `Authorization: Bearer <access_token>`.
- El login devuelve `{access_token, refresh_token, token_type, expires_in}` (segundos).

### Endpoints

#### Auth (`/api/v1/auth`)

| Método | Ruta | Descripción | Auth |
|--------|------|-------------|------|
| POST | `/auth/login` | Login con `{username, password}` | No |
| POST | `/auth/refresh` | Renueva sesión con `{refresh_token}` | No |

#### Categorías (`/api/v1/categories`)

| Método | Ruta | Descripción |
|--------|------|-------------|
| GET | `/categories` | Listado paginado |
| POST | `/categories` | Crear (`{name}`) |
| GET | `/categories/{id}` | Obtener |
| PATCH | `/categories/{id}` | Actualización parcial (`{name}`) |
| DELETE | `/categories/{id}` | Eliminar |

#### Productos (`/api/v1/products`)

Field `price` se envía/recibe en **decimal** (`10.84`) y se almacena en **céntimos** (`1084`).

| Método | Ruta | Descripción |
|--------|------|-------------|
| GET | `/products` | Listado paginado |
| POST | `/products` | Crear (ver ejemplo) |
| GET | `/products/{id}` | Obtener |
| PATCH | `/products/{id}` | Actualización parcial |
| DELETE | `/products/{id}` | Eliminar |

#### Roles (`/api/v1/roles`)

Roles válidos: `ADMIN`, `CUSTOMER`, `EDITOR`. CRUD completo (`{name}`), mismo patrón que categorías.

#### Usuarios (`/api/v1/users`)

| Método | Ruta | Descripción |
|--------|------|-------------|
| GET | `/users` | Listado paginado |
| POST | `/users` | Crear (`{email, password, username, role_id, photo?}`) |
| GET | `/users/{id}` | Obtener |
| PATCH | `/users/{id}` | Actualización parcial |
| DELETE | `/users/{id}` | Eliminar |

### PATCH parcial y campos nullables

Los PATCH actualizan solo los campos presentes. `img` (producto) y `photo` (usuario) son nullables con 3 estados:

- **Omitir el campo** → no se modifica.
- **`"img": null`** (o `""`) → se limpia a `NULL`.
- **Valor** → se actualiza el link.

### Ejemplos con curl

Login:

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123456"}'
```

```json
{
  "message": "Login successful",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "token_type": "Bearer",
    "expires_in": 1800
  },
  "error": null
}
```

Listar productos (paginado):

```bash
curl "http://localhost:8080/api/v1/products?page=1&page_size=10" \
  -H "Authorization: Bearer <access_token>"
```

Crear un producto:

```bash
curl -X POST http://localhost:8080/api/v1/products \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Laptop Gamer",
    "price": 1299.99,
    "category": 1,
    "in_stock": true,
    "quantity": 5,
    "img": "https://example.com/laptop.png"
  }'
```

Actualizar parcialmente (limpiar `img`):

```bash
curl -X PATCH http://localhost:8080/api/v1/products/1 \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{"img": null, "quantity": 8}'
```

Renovar sesión:

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "<refresh_token>"}'
```

## Modelo de datos

- **categories**: `id`, `name` (único).
- **products**: `id`, `name`, `price` (`BIGINT` en céntimos), `category` (FK → categories, `ON DELETE RESTRICT`), `in_stock` (default `false`), `quantity` (default `0`), `img` (nullable).
- **roles**: `id`, `name` (único). Roles sembrados: `ADMIN`, `CUSTOMER`, `EDITOR`.
- **users**: `id`, `email` (único), `password` (hash bcrypt), `username` (único), `role_id` (FK → roles), `photo` (nullable).

## Seguridad

- Passwords almacenados solo como hash bcrypt; nunca se devuelven en respuestas.
- JWT firmado con HS256 usando `JWT_SECRET`; no hay storage/revocación de tokens en servidor (para revocar, disminuir el TTL o rotar el secreto).
- `decodeJSON` limita el tamaño del cuerpo (1 MB) y rechaza campos desconocidos.#   e c o m e r c e - a p i - g o l a n g  
 