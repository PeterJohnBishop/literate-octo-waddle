# authentication-containerized

Project layout and entrypoints
- Main entry: main.go -> server.StartAuthenticationServer()
- HTTP: Gin router (server/router.go); protected routes under /api with JWT middleware
- DB: Postgres connector in db/postgres.go; CRUD in server/handlers/users.go
- JWT: middleware/jwt.go issues HS256 access/refresh tokens via env secrets

Environment and configuration
- Required env (loaded via github.com/joho/godotenv): PSQL_HOST, PSQL_PORT, PSQL_USER, PSQL_PASSWORD, PSQL_DBNAME, ACCESS_SECRET, REFRESH_SECRET
- PSQL_HOST=localhost for local testing
- .env is mandatory locally; do not commit secrets. In CI, provide via environment or secret store

# update

docker build -t peterjbishop/literate-octo-waddle:latest .
docker push peterjbishop/literate-octo-waddle:latest
docker pull peterjbishop/literate-octo-waddle:latest
docker-compose down 
docker-compose build --no-cache 
docker-compose up