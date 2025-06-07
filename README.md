# Bank

A Go-based banking application with PostgreSQL database.

## Prerequisites

- Go 1.24+
- Docker and Docker Compose (for database)

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

3. Set up environment variables (see Environment Configuration above)

4. Start the database:
   ```bash
   docker-compose up -d
   ```

## Test

### Run Tests
```bash
go test ./...
```

## Run Locally

1. Start the database:
   ```bash
   docker-compose up -d
   ```

2. Run the web application:
   ```bash
   with_env go run cmd/web/main.go
   ```

3. Run the worker (in another terminal):
   ```bash
   export $(cat .env | xargs)
   go run cmd/worker/main.go
   ```

The web application will be available at `http://localhost:80` (or the port specified in your environment).

## Development Workflow

1. **Database**: Always start with `docker-compose up -d`
2. **Environment**: Load variables with `export $(cat .env | xargs)`
3. **Development**: Run `go run cmd/web/main.go` for the web server
4. **Testing**: Run integration tests against the containerized database
5. **Cleanup**: Stop database with `docker-compose down`