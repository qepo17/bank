# Bank

A Go-based banking application with PostgreSQL database.

## Prerequisites

- Go 1.24+
- Docker and Docker Compose (for database)
- Make (for running development commands)
- [Goose](https://github.com/pressly/goose) (for database migrations)

## Environment Configuration

### Using Environment File

1. Copy the example environment file:
   ```bash
   cp example.env .env
   ```

2. Edit `.env` with your configuration

3. Load environment variables before running the application. You can put this function in .bashrc / .zshrc / fish config, then run `with_env <command>`:
   ```bash
   # Using export (bash/zsh)
   with_env () {
    eval "$(grep -vE '^#' "$1" | xargs)" "${@:2}"
   }
   
   # Or using fish
   function with_env
    set env_file '.env'
    set cmd $argv[2..-1]
    
    # Load environment variables from file, skipping comments
    for line in (grep -vE '^#' $env_file)
        set var_parts (string split '=' $line)
        if test (count $var_parts) -ge 2
            set -gx $var_parts[1] (string join '=' $var_parts[2..-1])
        end
    end
    
    # Execute the command
    eval $cmd
   end

   # Or import to vscode
   go.testEnvFile=${path}/.env
   ```

### Using Docker Compose for Database

Start PostgreSQL database:
```bash
docker-compose up -d
```

## Installation

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod download
   ```

3. Install goose (database migration tool):
   ```bash
   make install-goose
   ```

4. Set up environment variables (see Environment Configuration above)

5. Start the quick-setup command (database & dependencies & migration):
   ```bash
   make dev-setup
   ```

## Database Migrations

This project uses [Goose](https://github.com/pressly/goose) for database migrations.

### Migration Commands

#### Install Goose
```bash
make install-goose
```

#### Run Migrations
```bash
# Run all pending migrations
make migrate-up

# Check migration status
make migrate-status

# Rollback last migration
make migrate-down
```

#### Create New Migration
```bash
# Create a new migration file
make migrate-create name=add_user_table

# This will create a file like: migration/20240101120000_add_user_table.sql
```

#### Manual Goose Commands
If you prefer to use goose directly:
```bash
# Set up envar
GOOSE_DRIVER=postgres
GOOSE_DBSTRING=postgres://bankuser:bankpass@localhost:5432/bank
GOOSE_MIGRATION_DIR=./migration

# Run migrations
goose up

# Check status
goose status

# Create new migration
goose create create_user_table sql
```

## Test

### Run Tests
```bash
go test ./...
```

## Run Locally

1. Start the quick-setup:
   ```bash
   make dev-setup
   ```

2. Run the web application:
   ```bash
   make dev-run
   ```

3. Run the worker (in another terminal):
   ```bash
   make dev-worker
   ```

The web application will be available at `http://localhost:8080` (or the port specified in your environment).

## Development Workflow

1. **Database**: Always start with `docker-compose up -d`
2. **Environment**: Load variables with `with_env` (if you already specifiec as written above)
3. **Development**: Run `go run cmd/web/main.go` for the web server
4. **Testing**: Run integration tests against the containerized database
5. **Cleanup**: Stop database with `docker-compose down`