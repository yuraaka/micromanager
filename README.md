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
go install github.com/yuraaka/micromanager/cmd/mm@latest
```

Or download pre-built binaries from [releases](https://github.com/yuraaka/micromanager/releases).

## Quick Start

```bash
# Initialize a new project directory
mm init my-project

# Navigate to the project
cd my-project

# Create a new service
mm new api-service

# Run the service locally
mm run services/api-service
```

Your service is now running with the configuration from `service.toml`.

## Usage

```bash
# Initialize a new project directory
mm init <path> [--lang <language>]

# Create a new service within the project
mm new <service-name> [--empty]

# Run a service
mm run <path-to-service> [-m <mode>]

# Test services
mm test [path]

# Manage language packs
mm packs list
mm packs validate
```

### Commands

**init** - Initialize repository defaults and structure
- `--lang`: Default language for services (default: "go")

**new** - Create a new service skeleton
- `--empty`: Generate an external/empty service (Dockerfile + service.toml only)

**run** - Build and run a service with environment from service.toml
- `-m, --mode`: Environment mode: `local`, `docker`, or `minikube` (default: "local")

**test** - Run tests for all services or a specific service

**packs** - Manage language packs (list, validate)

### Examples

```bash
# Initialize a Go project
mm init my-project --lang go
cd my-project

# Create a new service
mm new auth-service

# Create an empty/external service (just Dockerfile + config)
mm new redis --empty

# Run service locally
mm run services/auth-service

# Run service in Docker
mm run services/auth-service -m docker

# Run service in Minikube
mm run services/auth-service -m minikube

# Test all services
mm test

# Test specific service
mm test services/auth-service

# List available language packs
mm packs list
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

### Testing & Deployment
- [ ] **Multi-environment testing**: Test services in local, Docker, and Minikube environments
- [ ] **Service orchestration**: Run and test multiple microservices together
- [ ] **Deployment pipeline generation**: Auto-generate GitHub Actions, GitLab CI, and other CI/CD workflows
- [ ] **Mock & stub generation**: Automatically generate mocks and stubs for advanced testing
- [ ] **Integration test scaffolding**: Pre-configured test environments for service dependencies

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

