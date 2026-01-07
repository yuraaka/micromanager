# MicroManager

A CLI tool for rapidly scaffolding production-ready microservices with best practices built-in.

## The Problem

Building microservices from scratch is repetitive and time-consuming. Developers spend hours setting up:

- Project structure and configuration
- Server boilerplate (HTTP/gRPC)
- Database connections and migrations
- Observability (logging, metrics, tracing)
- Containerization (Docker, docker-compose)
- Development workflows (Makefiles, hot-reload)

Every new service requires recreating this foundation, leading to inconsistency across teams and projects.

## The Solution

MicroManager generates fully-configured microservice projects from templates in seconds. Get a production-ready service with routing, middleware, database setup, Docker configuration, and development tools - all following best practices.

## Current Features

### Project Generation
- **Multi-language support**: Go (with more languages planned)
- **Service types**: REST APIs, gRPC services, workers
- **Smart templating**: Customizable project structure with variable substitution

### Go Services Include
- **HTTP Server**: Chi router with health checks and CORS
- **Database**: PostgreSQL integration with connection pooling
- **Observability**: Structured logging (slog), metrics endpoints
- **Configuration**: Environment-based config with validation
- **Containerization**: Multi-stage Dockerfile and docker-compose setup
- **Development Tools**: Makefile for common tasks, live reload with Air
- **Testing**: Test structure and examples

### Developer Experience
- **Fast setup**: Generate a complete service in under 10 seconds
- **Opinionated defaults**: Battle-tested configurations out of the box
- **Customizable**: Templates can be modified to match your conventions
- **Zero dependencies**: Single binary, no runtime requirements

## Installation

```bash
go install github.com/yuraaka/micromanager/cmd/micromanager@latest
```

Or download pre-built binaries from [releases](https://github.com/yuraaka/micromanager/releases).

## Quick Start

```bash
# Create a new Go REST API service
micromanager new my-service --lang go --type service

# Navigate to the project
cd my-service

# Start development environment
make dev
```

Your service is now running at `http://localhost:8080` with hot-reload enabled.

## Usage

```bash
micromanager new <project-name> [flags]
```

### Flags
- `--lang`: Programming language (default: "go")
- `--type`: Service type (default: "service")
- `--output`: Output directory (default: current directory)

### Examples

```bash
# Create a Go service with custom output path
micromanager new my-api --lang go --output ./services/my-api

# Create a worker service
micromanager new my-worker --type worker
```

## Project Structure (Go)

```
my-service/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── handlers/
│   │   └── handlers.go          # HTTP handlers
│   └── middleware/
│       └── middleware.go        # HTTP middleware
├── .air.toml                    # Live reload configuration
├── docker-compose.yml           # Local development stack
├── Dockerfile                   # Multi-stage production build
├── Makefile                     # Development commands
└── go.mod
```

## Future Features

### Language Support
- [ ] Python (FastAPI)
- [ ] Node.js (Express/Fastify)
- [ ] Rust (Actix/Axum)
- [ ] Java (Spring Boot)

### Service Types
- [ ] GraphQL APIs
- [ ] WebSocket services
- [ ] Event-driven consumers (Kafka, RabbitMQ)
- [ ] Scheduled jobs/cron workers
- [ ] CLI applications

### Advanced Features
- [ ] **Multi-service projects**: Generate entire microservice ecosystems
- [ ] **Service mesh integration**: Istio, Linkerd configurations
- [ ] **API Gateway**: Auto-generate gateway configurations
- [ ] **Authentication**: JWT, OAuth2 templates
- [ ] **CI/CD**: GitHub Actions, GitLab CI, Jenkins pipelines
- [ ] **Infrastructure as Code**: Terraform, Kubernetes manifests
- [ ] **Message queuing**: RabbitMQ, Kafka, NATS integration
- [ ] **Caching**: Redis, Memcached setups
- [ ] **Custom templates**: Import and share community templates
- [ ] **Interactive mode**: Guided project setup with prompts
- [ ] **Migration tools**: Convert existing projects to MicroManager structure

### Developer Experience
- [ ] Web UI for template management
- [ ] VS Code extension
- [ ] Project upgrade command (update generated files)
- [ ] Template marketplace
- [ ] Analytics and metrics tracking

## Configuration

Templates are located in `pack/lang/<language>/templates/`. You can customize them or add your own.

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Why MicroManager?

**Consistency**: Every service follows the same patterns and best practices.

**Speed**: Go from idea to running service in minutes, not hours.

**Maintainability**: Well-structured projects are easier to understand and modify.

**Learning**: New team members can study generated code to learn your standards.

**Focus**: Spend time building features, not setting up infrastructure.

---

