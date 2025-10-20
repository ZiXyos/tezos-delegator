# Tezos Delegation Indexer

A Go-based service that continuously indexes Tezos delegation data from the TzKT API and provides a REST API for querying delegation information.

## 🚀 Features

- **Continuous Indexing**: Polls Tezos delegations from TzKT API and stores them in PostgreSQL
- **REST API**: Exposes delegation data via HTTP endpoints
- **Historical Data**: Supports backfilling and incremental updates
- **Health Monitoring**: Built-in health checks and observability
- **Docker Support**: Containerized deployment with Docker Compose

## 📋 Prerequisites

### Required Dependencies

- **Go 1.25.1+** - For local development
- **Docker & Docker Compose** - For containerized deployment
- **PostgreSQL 15+** - Database backend (included in docker-compose)

## 🛠️ Installation & Setup

### Option 1: Docker Compose (Recommended)

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd delegator
   ```

2. **Start the services**
   ```bash
   docker-compose up -d
   ```

   This will start:
   - **Delegator Service** on port `8888`
   - **PostgreSQL Database** on port `54323`
   - **Adminer** (Database UI) on port `8881`

3. **Verify the service is running**
   ```bash
   curl http://localhost:8888/health
   ```

### Option 2: Local Development

1. **Install Go 1.25.1+**
   ```bash
   # Check Go version
   go version
   ```

2. **Install dependencies**
   ```bash
   go mod download
   go mod verify
   ```

3. **Start PostgreSQL** (using Docker)
   ```bash
   docker-compose up postgres -d
   ```

4. **Set environment variables**
   ```bash
   export GO_ENV=local
   export APP_ENV=local
   export PORT=8888
   export CONFIG_PATH=./conf/config.local.toml
   ```

5. **Run the application**
   ```bash
   go run .
   ```

## 🎯 Usage

### API Endpoints

#### Health Check
```bash
GET /health
```
**Response:**
```json
{
  "data": "ok"
}
```

#### Get Delegations
```bash
GET /xtz/delegations
```
**Response:**
```json
{
  "data": [
    {
      "timestamp": "2023-01-01T12:00:00Z",
      "amount": 100000,
      "delegator": "tz1...",
      "level": 1000
    }
  ]
}
```

### Configuration

The service uses TOML configuration files located in the `conf/` directory:

```toml
[service]
name = "delegator"
version = "1.0.0"

[http]
port = 8888
read_timeout = 3600
write_timeout = 3600

[storage]
    [storage.database]
    host = "postgres"
    port = 5432
    username = "delegator"
    database = "delegator_local"
    password = "password"

[logging]
level = "info"
format = "json"
```

## 🧪 Testing

### Run All Tests
```bash
go test ./... -v
```

### Run Tests with Coverage
```bash
go test ./... -cover
```

### Run Specific Package Tests
```bash
# Test core delegator functionality
go test ./internal/core/delegator/... -v

# Test HTTP services
go test ./internal/httpservice/... -v

# Test external services
go test ./internal/services/... -v
```

## 🐳 Docker

### Build Image
```bash
docker build -t delegator:latest \
  --build-arg TARGET_OS=linux \
  --build-arg TARGET_ARCH=amd64 \
  .
```

### Run Container
```bash
docker run -p 8888:8888 \
  -e GO_ENV=local \
  -e APP_ENV=local \
  delegator:latest
```

## 🔧 Development

### Project Structure
```
delegator/
├── cmd/                    # Application entrypoints
├── conf/                   # Configuration files
│   └── config.local.toml
├── internal/
│   ├── core/
│   │   └── delegator/      # Core business logic
│   ├── httpservice/        # HTTP server and routes
│   ├── services/           # External service clients
│   └── database/           # Database connections
├── pkg/
│   └── domain/             # Domain models and interfaces
├── mocks/                  # Generated mocks for testing
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── go.sum
```

### Key Dependencies

#### Core Dependencies
- **gin-gonic/gin** `v1.11.0` - HTTP web framework
- **gorm.io/gorm** `v1.31.0` - ORM for database operations
- **gorm.io/driver/postgres** `v1.6.0` - PostgreSQL driver
- **google/uuid** `v1.6.0` - UUID generation
- **golang-migrate/migrate** `v4.19.0` - Database migrations

#### Testing Dependencies
- **stretchr/testify** `v1.11.1` - Testing framework
- **uber.org/mock** `v0.5.0` - Mock generation

#### Configuration
- **zixyos/goloader** `v0.2.0` - Configuration loader
- **zixyos/glog** `v0.1.0` - Logging utilities

### Generate Mocks
```bash
# Install mockery if not already installed
go install github.com/vektra/mockery/v2@latest

# Generate mocks
mockery
```

### Database Migrations
```bash
# Migrations are handled automatically on startup
# Check internal/database/ for migration files
```

## 🚦 Health Checks

The service includes built-in health monitoring:

### Docker Health Check
```bash
# Check container health
docker-compose ps

# Manual health check
wget --no-verbose --tries=3 --spider http://localhost:8888/health
```

### Database Health
```bash
# Check PostgreSQL connection
docker-compose exec postgres pg_isready -U delegator -d delegator_local
```

## 📊 Monitoring

### Service Logs
```bash
# View service logs
docker-compose logs -f delegator

# View database logs
docker-compose logs -f postgres
```

### Database Administration
Access Adminer at `http://localhost:8881` with:
- **System**: PostgreSQL
- **Server**: postgres
- **Username**: delegator
- **Password**: password
- **Database**: delegator_local

## 🔨 Build & Deployment

### Production Build
```bash
# Build optimized binary
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
go build -ldflags='-w -s' -o delegator ./
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `GO_ENV` | Go environment | `local` |
| `APP_ENV` | Application environment | `local` |
| `PORT` | HTTP server port | `8888` |
| `CONFIG_PATH` | Path to config file | `/app/conf/config.local.toml` |

## 🔍 Troubleshooting

### Common Issues

1. **Port already in use**
   ```bash
   # Check what's using the port
   lsof -i :8888
   
   # Use different port
   PORT=8889 docker-compose up
   ```

2. **Database connection issues**
   ```bash
   # Check PostgreSQL logs
   docker-compose logs postgres
   
   # Restart database
   docker-compose restart postgres
   ```

3. **Configuration not found**
   ```bash
   # Ensure config file exists
   ls -la conf/config.local.toml
   
   # Check file permissions
   chmod 644 conf/config.local.toml
   ```

### Debug Mode
```bash
# Run with debug logging
GIN_MODE=debug docker-compose up
```

## 📝 License

[Add your license information here]

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📬 Support

For support and questions:
- Create an issue in the repository
- Check existing documentation
- Review logs for error details