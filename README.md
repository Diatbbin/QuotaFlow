# QuotaFlow

Go REST API for managing per-user AI tool token quotas and transferring spare tokens between users

Each user can register AI tools (e.g. ChatGPT) with a token limit. Users can transfer unused tokens to another user's tool of the **same type** (e.g. from ChatGPT to ChatGPT)

## Tech stack

- Go, Gin
- PostgreSQL, sqlc, golang-migrate
- PASETO, bcrypt
- Redis, Asynq (async transfer email notifications)
- Gmail SMTP
- Fly.io (deployment)
- testify

## Prerequisites for curl demo (testing the live app)

- [jq](https://jqlang.org/) (for readable curl output)

Install jq (on Mac):

```bash
brew install jq
```

## API endpoints

All routes except `POST /users` and `POST /users/login` require the user to be logged in


| Method | Path                        | Description                                                                                                  |
| ------ | --------------------------- | ------------------------------------------------------------------------------------------------------------ |
| POST   | `/users`                    | Register user: `username` (alphanumeric, min 3), `email`, `password` (min 8)                                 |
| POST   | `/users/login`              | Login: `username` (alphanumeric, min 3), `password` (min 8)                                                  |
| POST   | `/ai-tools`                 | Create AI tool (`token_limit` min 1)                                                                         |
| GET    | `/ai-tools`                 | List your AI tools, query using `page_id` (min 1), `page_size` (5–10)                                        |
| GET    | `/ai-tools/:id`             | Get a specific AI tool owned by the current user                                                             |
| PUT    | `/ai-tools/:id`             | Update `token_limit` (min 0)                                                                                 |
| PUT    | `/ai-tools/:id/tokens-used` | Update `tokens_used` (min 0)                                                                                 |
| DELETE | `/ai-tools/:id`             | Delete AI tool                                                                                               |
| POST   | `/token-transfers`          | Transfer spare tokens (min 1), using `from_ai_tool_id`, `to_ai_tool_id`, `amount of tokens to be transferred` |


## Quick demo (live app)

No local setup required. Run these commands in your terminal

### Set variables

Username must be alphanumeric (min 3 characters). Password must be at least 8 characters.

Set `USER2_EMAIL` to **your own email** — user2 should receive the transfer notification after a successful token transfer (requires email/Redis configured on the deployed app)

```bash
BASE_URL="https://quotaflow.fly.dev"

USER1_EMAIL="user1@example.com"
USER2_EMAIL="user2@example.com" # -> Use your email to receive transfer notifications

USER1="user1"
USER2="user2"

USER1_PW="password123"
USER2_PW="password456"
```

If `user1` / `user2` already exist from a previous run, pick different usernames **and** emails — both must be unique (e.g. `user3` / `user3@example.com`, `user4` / `user4@example.com`)

### Register two users

```bash
curl --max-time 90 -s -X POST "$BASE_URL/users" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USER1\",\"email\":\"$USER1_EMAIL\",\"password\":\"$USER1_PW\"}" | jq .

curl --max-time 90 -s -X POST "$BASE_URL/users" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USER2\",\"email\":\"$USER2_EMAIL\",\"password\":\"$USER2_PW\"}" | jq .
```

### Login

```bash
USER1_TOKEN=$(curl --max-time 90 -s -X POST "$BASE_URL/users/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USER1\",\"password\":\"$USER1_PW\"}" | jq -r '.access_token')

if [ -n "$USER1_TOKEN" ] && [ "$USER1_TOKEN" != "null" ]; then
    echo "user logged in successfully"
fi

USER2_TOKEN=$(curl --max-time 90 -s -X POST "$BASE_URL/users/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USER2\",\"password\":\"$USER2_PW\"}" | jq -r '.access_token')

if [ -n "$USER2_TOKEN" ] && [ "$USER2_TOKEN" != "null" ]; then
    echo "user logged in successfully"
fi
```

### Create AI tools

Both users need a tool with the **same type** (e.g. `chatgpt`). The owner is taken from your access token — do not send `username` in the body.

```bash
USER1_AI_TOOL_ID=$(curl --max-time 90 -s -X POST "$BASE_URL/ai-tools" \
  -H "Authorization: bearer $USER1_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"tool":"chatgpt","token_limit":1000}' \
  | jq -r '.ai_tool.id')

if [ -n "$USER1_AI_TOOL_ID" ] && [ "$USER1_AI_TOOL_ID" != "null" ]; then
    echo "AI tool with ID $USER1_AI_TOOL_ID created successfully"
fi

USER2_AI_TOOL_ID=$(curl --max-time 90 -s -X POST "$BASE_URL/ai-tools" \
  -H "Authorization: bearer $USER2_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"tool":"chatgpt","token_limit":100}' \
  | jq -r '.ai_tool.id')

if [ -n "$USER2_AI_TOOL_ID" ] && [ "$USER2_AI_TOOL_ID" != "null" ]; then
    echo "AI tool with ID $USER2_AI_TOOL_ID created successfully"
fi
```

### List tools

```bash
curl --max-time 90 -s "$BASE_URL/ai-tools?page_id=1&page_size=5" \
  -H "Authorization: bearer $USER1_TOKEN" | jq .
```

### Get a specific tool

```bash
curl --max-time 90 -s "$BASE_URL/ai-tools/$USER1_AI_TOOL_ID" \
  -H "Authorization: bearer $USER1_TOKEN" | jq .
```

### Transfer tokens

#### Transfer rules

- Sender must own the source AI tool
- Cannot transfer to yourself
- Both tools must be the same type (e.g. both `chatgpt`)
- Sender must have enough spare tokens (`token_limit - tokens_used`)

user1 sends 50 spare tokens to user2. After a successful transfer, user2 should receive an email at `$USER2_EMAIL`.

```bash
curl --max-time 90 -s -X POST "$BASE_URL/token-transfers" \
  -H "Authorization: bearer $USER1_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"from_ai_tool_id\":$USER1_AI_TOOL_ID,\"to_ai_tool_id\":$USER2_AI_TOOL_ID,\"tokens\":50}" | jq .
```

### Update a tool

Update token limit (e.g. 800):

```bash
curl --max-time 90 -s -X PUT "$BASE_URL/ai-tools/$USER1_AI_TOOL_ID" \
  -H "Authorization: bearer $USER1_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"token_limit":800}' | jq .
```

Update tokens used (e.g. 250):

```bash
curl --max-time 90 -s -X PUT "$BASE_URL/ai-tools/$USER1_AI_TOOL_ID/tokens-used" \
  -H "Authorization: bearer $USER1_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"tokens_used":250}' | jq .
```

### Delete a tool

```bash
curl --max-time 90 -s -X DELETE "$BASE_URL/ai-tools/$USER1_AI_TOOL_ID" \
  -H "Authorization: bearer $USER1_TOKEN" | jq .
```

## Running Unit and Integration Tests locally

Prerequisites: 

Go 1.26+

Docker 

[golang-migrate](https://github.com/golang-migrate/migrate)

### 1. Set up database

```bash
make postgres
make create-db   # If the database doesn't exist yet
make migrate-up
```

If containers already exist:

```bash
docker start postgres18
```

### 2. Configure environment

Create `app.env` in the project root (gitignored):

```env
DB_DRIVER=postgres
DB_SOURCE=postgres://root:test@localhost:5432/QuotaFlow?sslmode=disable
EMAIL_SENDER_NAME=QuotaFlow
EMAIL_SENDER_ADDR=your-gmail@gmail.com
EMAIL_PASSWORD=your-gmail-app-password
```

Use Gmail for EMAIL_SENDER_ADDR and [Gmail's app password](https://myaccount.google.com/apppasswords) for `EMAIL_PASSWORD`

### 3. Running tests

Requires:

- Postgres container running (`make postgres`)
- Database created and migrated (`make create-db`, `make migrate-up`)
- `app.env` configured in the project root

```bash
make test
```

