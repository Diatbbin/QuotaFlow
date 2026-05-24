# QuotaFlow

Go REST API for managing per-user AI tool token quotas and transferring spare tokens between users.

Each user can register AI tools (e.g. ChatGPT) with a token limit. Users can transfer unused tokens to another user's tool of the **same type** (e.g. from ChatGPT to ChatGPT).

## Tech stack

- Go, Gin
- PostgreSQL, sqlc, golang-migrate
- PASETO, bcrypt 
- testify

## Prerequisites

- Go 1.26+
- Docker
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI
- [jq](https://jqlang.org/) (for readable curl output)

## Setup

### 1. Set up database

```bash
make postgres 
make create-db  # If the database doesn't exist yet
make migrate-up
```

If the `postgres18` container already exists, run 

```bash
docker start postgres18
```

instead of 

```bash
make postgres
```

To reset an existing database, run 

```bash
make migrate-down
make migrate-up
```

### 2. Configure environment

Create `app.env` in the project root (gitignored):

Add the following env variables to app.env

```env
DB_DRIVER=postgres
DB_SOURCE=postgres://root:test@localhost:5432/QuotaFlow?sslmode=disable
SERVER_ADDRESS=0.0.0.0:8080
TOKEN_SYMMETRIC_KEY=12345678901234567890123456789012
ACCESS_TOKEN_DURATION=15m
```

`TOKEN_SYMMETRIC_KEY` must have 32 characters.

## API endpoints


| Method | Path                        | Description                                                                                                        |
| ------ | --------------------------- | ------------------------------------------------------------------------------------------------------------------ |
| POST   | `/users`                    | Register new user, username (alphanumeric & at least chars), password (at least 8 chars)                           |
| POST   | `/users/login`              | Login user, username (alphanumeric & at least chars), password (at least 8 chars)                                  |
| POST   | `/ai-tools`                 | Create AI tool for logged-in user (owner from access token; body: `tool`, `token_limit`), token_limit (at least 1) |
| GET    | `/ai-tools`                 | Retrieves your AI tools (paginated), page size must be (5-10)                                                      |
| GET    | `/ai-tools/:id`             | Get a specific AI tool                                                                                             |
| PUT    | `/ai-tools/:id`             | Update token limit                                                                                                 |
| PUT    | `/ai-tools/:id/tokens-used` | Update tokens used                                                                                                 |
| DELETE | `/ai-tools/:id`             | Delete AI tool                                                                                                     |
| POST   | `/token-transfers`          | Transfer spare tokens (at least 1)                                                                                 |


### Demo

Tip: pipe curl output through `jq .` to format the JSON output properly. Without it, responses appear as one long line

Install jq (on Mac) : 

```bash
brew install jq
```

### In terminal, run

```bash
make server
```

API listens on `http://localhost:8080`.

### On another terminal, register two users

Username must be alphanumeric (min 3 characters). Password must be at least 8 characters.

- **user1**: `user1@example.com` / `password123`
- **user2**: `user2@example.com` / `password456`

```bash
curl -s -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"username":"user1","email":"user1@example.com","password":"password123"}' | jq .

curl -s -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"username":"user2","email":"user2@example.com","password":"password456"}' | jq .
```

### Login the two users created earlier

```bash
USER1_TOKEN=$(curl -s -X POST http://localhost:8080/users/login \
  -H "Content-Type: application/json" \
  -d '{"username":"user1","password":"password123"}' | jq -r '.access_token')

if [ -n "$USER1_TOKEN" ] && [ "$USER1_TOKEN" != "null" ]; then
    echo "user logged in successfully"
fi

USER2_TOKEN=$(curl -s -X POST http://localhost:8080/users/login \
  -H "Content-Type: application/json" \
  -d '{"username":"user2","password":"password456"}' | jq -r '.access_token')

if [ -n "$USER2_TOKEN" ] && [ "$USER2_TOKEN" != "null" ]; then
    echo "user logged in successfully"
fi
```

### Create AI tools for each user so that tokens can be transferred later

Both users need a tool with the **same type** (e.g. `chatgpt`). The owner is taken from your access token—do not send `username` in the body.

```bash
USER1_AI_TOOL_ID=$(curl -s -X POST http://localhost:8080/ai-tools \
  -H "Authorization: bearer $USER1_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"tool":"chatgpt","token_limit":1000}' \
  | jq -r '.ai_tool.id')

if [ -n "$USER1_AI_TOOL_ID" ] && [ "$USER1_AI_TOOL_ID" != "null" ]; then
    echo "AI tool with ID $USER1_AI_TOOL_ID created successfully"
fi

USER2_AI_TOOL_ID=$(curl -s -X POST http://localhost:8080/ai-tools \
  -H "Authorization: bearer $USER2_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"tool":"chatgpt","token_limit":100}' \
  | jq -r '.ai_tool.id')

if [ -n "$USER2_AI_TOOL_ID" ] && [ "$USER2_AI_TOOL_ID" != "null" ]; then
    echo "AI tool with ID $USER2_AI_TOOL_ID created successfully"
fi
```

### List your tools

Each user only sees their own tools. Using page_id of 1 and page_size of 5 (must be 5-10)

```bash
curl -s "http://localhost:8080/ai-tools?page_id=1&page_size=5" \
  -H "Authorization: bearer $USER1_TOKEN" | jq .
```

### Transfer tokens

#### Transfer rules

- Sender must own the source AI tool
- Cannot transfer to yourself
- Both tools must be the same type (e.g. both `chatgpt`)
- Sender must have enough spare tokens (`token_limit - tokens_used`)

user1 sends 50 spare tokens to user2:

```bash
curl -s -X POST http://localhost:8080/token-transfers \
  -H "Authorization: bearer $USER1_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"from_ai_tool_id\":$USER1_AI_TOOL_ID,\"to_ai_tool_id\":$USER2_AI_TOOL_ID,\"tokens\":50}" | jq .
```

### Update a tool

Update token limit (e.g. 800):

```bash
curl -s -X PUT http://localhost:8080/ai-tools/$USER1_AI_TOOL_ID \
  -H "Authorization: bearer $USER1_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"token_limit":800}' | jq .
```

Update tokens used (e.g. 250):

```bash
curl -s -X PUT http://localhost:8080/ai-tools/$USER1_AI_TOOL_ID/tokens-used \
  -H "Authorization: bearer $USER1_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"tokens_used":250}' | jq .
```

### Delete a tool

Delete only works for tools you own.

Deleting the ChatGPT tool created earlier for user1:

```bash
curl -s -X DELETE http://localhost:8080/ai-tools/$USER1_AI_TOOL_ID \
  -H "Authorization: bearer $USER1_TOKEN" | jq .
```

## Running Tests

```bash
make test
```

