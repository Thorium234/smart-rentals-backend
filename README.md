git pull

# Prereqisities
install go 1.24.11+

# Install required dependencies
```bash
  go download
```
# Run your server
```go
go run cmd/server/main.go
```
##***TROUBLESHOOTING***
-must have .env since i'll not push
```shell
touch backend-gin/.env
```
add these to .env file
```go
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

DB_HOST=localhost
DB_PORT=5432
DB_USER=
DB_PASSWORD=
DB_NAME=

DB_SSLMODE=disable
JWT_SECRET=
ENV=development
```
**to generate JWT_SECRET**
```shell
openssl rand -base64 64
```

# Database migrations
-install the migrate cli
```shell
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```
-run the migration
```shell
migrate -database "postgres://zolet:test@localhost:5432/auth_service?sslmode=disable" -path migrations up
```

# Database connection and creation
**
# Api testing
```shell
curl 
```
