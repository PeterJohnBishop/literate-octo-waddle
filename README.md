# 

Project layout and entrypoints
- Main entry: main.go -> server.StartAuthenticationServer()
- HTTP: Gin router (server/router.go); protected routes under /api with JWT middleware
- DB: Postgres connector in db/postgres.go; CRUD in server/handlers/users.go
- JWT: middleware/jwt.go issues HS256 access/refresh tokens via env secrets

Environment and configuration
- Required env (loaded via github.com/joho/godotenv): PSQL_HOST, PSQL_PORT, PSQL_USER, PSQL_PASSWORD, PSQL_DBNAME, ACCESS_SECRET, REFRESH_SECRET
- PSQL_HOST=localhost for local testing
- .env is mandatory locally; do not commit secrets. In CI, provide via environment or secret store

# API Summary

This is a JWT-based authentication and user management API with websocket connection
Base URL: http://localhost:8081 or http://localhost with Nginx

## Authentication

POST /auth/login
Login with email & password.
Body:
{
  "email": "string",
  "password": "string"
}
Responses: 200 AuthResponse | 400 Invalid | 401 Unauthorized | 404 User not found
POST /auth/register
Register a new user.
Body:
{
  "name": "string",
  "email": "string",
  "password": "string"
}
Responses: 201 Created | 400 Invalid | 500 Server error
GET /auth/refresh
Refresh access token.
Body:
{
  "refresh_token": "string"
}
Responses: 200 RefreshResponse | 400 Invalid | 401 Invalid token
POST /auth/logout
Logout a user.
Body:
{
  "id": "string"
}
Responses: 200 Logged out | 400 Invalid

## Users (JWT Required)

GET /api/users
List all users.
Responses: 200 Array of User | 401 Unauthorized
PUT /api/users
Update a user.
Body:
{
  "id": "string",
  "name": "string",
  "email": "string",
  "password": "string",
  "online": true,
  "files": ["string"]
}
Responses: 200 Updated | 401 Unauthorized | 404 Not found
GET /api/users/{id}
Get user by ID.
Responses: 200 User | 401 Unauthorized | 404 Not found
DELETE /api/users/{id}
Delete user by ID.
Responses: 200 Deleted | 401 Unauthorized | 404 Not found
POST /api/users/password
Update password.
Body:
{
  "userId": "string",
  "currentPassword": "string",
  "newPassword": "string"
}
Responses: 200 Updated | 400 Invalid | 401 Unauthorized | 404 Not found

## websocket

ws://localhost/ws?token={token}

## Schemas

User
{
  "id": "string",
  "name": "string",
  "email": "string",
  "password": "string (hashed)",
  "online": "boolean",
  "files": ["string"],
  "created": 123456789,
  "updated": 123456789
}

AuthResponse
{
  "message": "string",
  "token": "jwt",
  "refresh_token": "jwt"
}

RefreshResponse
{
  "token": "jwt"
}

# update

docker build -t peterjbishop/literate-octo-waddle:latest .
docker push peterjbishop/literate-octo-waddle:latest
docker pull peterjbishop/literate-octo-waddle:latest
docker-compose down 
docker-compose build --no-cache 
docker-compose up