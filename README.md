<div align="center">
  <img src="https://github.com/user-attachments/assets/995be885-ec69-4d17-b177-8105eb7c09da" height="350" alt="VolcAPI" />
  <p><strong>OpenAPI-Native API Testing Tool Built in Go</strong></p>
</div>

**VolcAPI** is a modern, configuration-driven API testing tool that makes your OpenAPI specifications executable. Define your tests once in your OpenAPI spec, run them anywhere—local development, CI/CD pipelines, or production monitoring.

At its core, VolcAPI transforms static API documentation into living, automated test suites. Whether you're a backend developer validating endpoints or a DevOps engineer setting up CI/CD, VolcAPI provides the tools to ensure your APIs work as expected.

This is **v0.1.0-alpha**, the foundation for a comprehensive API testing platform that aims to unify functional testing, performance testing, and monitoring.

This tool is built for developers who want fast, reliable API testing without the complexity of multiple tools.

---

## 🎯 Why VolcAPI?

Development teams currently struggle with:
- **Tool Fragmentation**: Using Postman for functional tests, K6 for performance, separate monitoring tools
- **Double Maintenance**: OpenAPI specs for documentation, separate test configs for validation
- **Poor CI/CD Integration**: Existing tools don't fit modern development workflows
- **Environment Management**: Manual config switching between local/staging/production

**VolcAPI solves this** by making your OpenAPI specification the single source of truth for both documentation and testing.

---

## 🚀 Key Features

### ✅ Currently Available (v0.1.0-alpha)
1. **OpenAPI-Native Testing**
   - Parse OpenAPI 3.x specifications
   - Define test scenarios directly in your spec using `v-functional-test` extensions
   - Auto-validate requests and responses against your API schema

2. **Flexible Scenario Management**
   - Define reusable test scenarios in OpenAPI spec or separate config files
   - Support for headers, query parameters, request bodies
   - Response validation with JSON matching and contains checks

3. **Environment Configuration**
   - Separate config files for different environments (`volcapi_local.yml`, `volcapi_staging.yml`)
   - Environment variable support
   - Custom host URLs per environment

4. **Developer-Friendly CLI**
   - Simple command structure
   - Support for local files and remote URLs
   - Clean, readable output

### 🚧 Coming Soon
- **Performance Testing**: Load testing with configurable scenarios
- **Monitoring**: Continuous API health checks with alerting
- **Advanced Validations**: Schema validation, regex patterns, type checking
- **Multiple Output Formats**: JSON, JUnit XML for CI/CD integration
- **Web Dashboard**: Historical results and team collaboration
- **Integrations**: Grafana, Slack, GitHub Actions

---

## 📖 Getting Started

### Prerequisites

Make sure you have **Go 1.21+** installed:
```bash
go version
```

### Installation

**Option 1: Install from source**
```bash
git clone https://github.com/volcapi/volcapi.git
cd volcapi
go build -o volcapi
sudo mv volcapi /usr/local/bin/
```

**Option 2: Go install (when available)**
```bash
go install github.com/volcapi/volcapi@latest
```

### Quick Start

1. **Create your OpenAPI spec with test scenarios:**

```yaml
# openapi.yml
openapi: 3.0.3
info:
  title: My API
  version: 1.0.0
servers:
  - url: https://api.example.com

# Define reusable scenarios
scenarios:
  valid_login:
    headers:
      Content-Type: application/json
    request:
      email: user@example.com
      password: password123
    response:
      status: 200
      body:
        contains: ["token", "user"]

paths:
  /auth/login:
    post:
      summary: User login
      responses:
        '200':
          description: Success
      v-functional-test:
        scenarios: ["valid_login"]
```

2. **Create environment config:**

```yaml
# volcapi_local.yml
host: http://localhost:3000

scenarios:
  # Additional local-specific scenarios
  test_user:
    headers:
      Authorization: Token ${TOKEN}
    response:
      status: 200

env:
  TOKEN: your_local_token_here
```

3. **Run your tests:**

```bash
# Run tests with OpenAPI spec
volcapi run volcapi_local.yml -o openapi.yml
```

---

## 📚 Configuration Guide

### OpenAPI Extensions

VolcAPI uses custom OpenAPI extensions to define test scenarios:

```yaml
paths:
  /api/users/{id}:
    get:
      summary: Get user by ID
      responses:
        '200':
          description: Success
      
      # VolcAPI test scenarios
      v-functional-test:
        scenarios: ["get_valid_user", "get_invalid_user"]

# Define scenarios at the root level
scenarios:
  get_valid_user:
    query:
      id: 123
    headers:
      Accept: application/json
    response:
      status: 200
      body:
        json:
          id:
            value: 123
          email:
            exists: true

  get_invalid_user:
    query:
      id: 99999
    response:
      status: 404
```

### Environment Configuration

```yaml
# volcapi_local.yml
host: http://localhost:3000

# Environment-specific scenarios
scenarios:
  auth_test:
    headers:
      Authorization: Bearer ${API_TOKEN}
    response:
      status: 200

# Environment variables
env:
  API_TOKEN: your_token_here
  MAX_TIMEOUT: "30"
```

### Response Validation

**Check specific values:**
```yaml
response:
  status: 200
  body:
    json:
      user:
        object:
          name:
            value: "John Doe"
          email:
            value: "john@example.com"
```

**Check field existence:**
```yaml
response:
  body:
    contains: ["user_id", "email", "token"]
```

**Validate nested objects:**
```yaml
response:
  body:
    json:
      headers:
        object:
          Accept:
            value: "application/json"
          Host:
            value: "api.example.com"
```

---

## 🔧 CLI Reference

### Commands

```bash
# Run tests from config file
volcapi run <config-path> [flags]

# Run with OpenAPI spec
volcapi run volcapi_local.yml -o openapi.yml

# Run from remote URL
volcapi run https://example.com/volcapi_local.yml -o openapi.yml
```

### Flags

- `-o, --openapi <path>`: Path to OpenAPI specification file
- `-h, --help`: Show help for commands

---

## 💡 Examples

### Example 1: Simple GET Request

```yaml
# openapi.yml
paths:
  /get:
    get:
      summary: Echo request
      responses:
        '200':
          description: OK
      v-functional-test:
        scenarios: ["simple_get"]

scenarios:
  simple_get:
    headers:
      Accept: application/json
    response:
      status: 200
      body:
        contains: ["headers.Host"]
```

### Example 2: GET with Query Parameters

```yaml
scenarios:
  get_with_query:
    query:
      id: 132
      filter: active
    headers:
      Accept: application/json
    response:
      status: 200
      body:
        json:
          args:
            object:
              id:
                value: "132"
              filter:
                value: "active"
```

### Example 3: POST with Authentication

```yaml
scenarios:
  create_user:
    headers:
      Content-Type: application/json
      Authorization: Bearer ${TOKEN}
    request:
      name: "Alice"
      email: "alice@example.com"
    response:
      status: 201
      body:
        contains: ["id", "name", "email"]

env:
  TOKEN: your_api_token
```

---

## 🚀 CI/CD Integration

### GitHub Actions (Coming Soon)

```yaml
name: API Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Install VolcAPI
        run: go install github.com/yourusername/volcapi@latest
      
      - name: Run API Tests
        env:
          STAGING_TOKEN: ${{ secrets.STAGING_TOKEN }}
        run: volcapi run volcapi_staging.yml -o openapi.yml
```

---

## 🛣️ Roadmap

### ✅ Phase 1 (Current - v0.1.0-alpha)
- [x] OpenAPI 3.x parsing
- [x] Basic functional testing (GET requests)
- [x] Environment configuration system
- [x] Response validation (status, JSON matching)
- [x] CLI tool with basic commands
- [ ] POST/PUT/DELETE request support
- [ ] Advanced response validation (regex, types)
- [ ] Better error reporting

### 🚧 Phase 2 (Next 2-3 months)
- [ ] Performance/load testing
- [ ] Multiple output formats (JSON, JUnit)
- [ ] Web dashboard for results
- [ ] Grafana integration
- [ ] GitHub Actions marketplace

### 🔮 Phase 3 (6+ months)
- [ ] Continuous monitoring
- [ ] Slack/Discord alerts
- [ ] Team collaboration features
- [ ] Advanced analytics
- [ ] OpenAPI 3.1 support

---

## 🤝 Contributing

Contributions are welcome! Here's how you can help:

1. **Report Bugs**: Open an issue with details and reproduction steps
2. **Suggest Features**: Share your ideas in GitHub Issues
3. **Submit PRs**: Fork the repo, make changes, submit a pull request
4. **Improve Docs**: Help make documentation clearer

### Development Setup

```bash
# Clone the repository
git clone https://github.com/yourusername/volcapi.git
cd volcapi

# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build -o volcapi

# Run locally
./volcapi run volcapi_local.yml -o openapi.yml
```

---

## 📝 License

MIT License - see [LICENSE](LICENSE) file for details

---

## 🌟 Show Your Support

If VolcAPI helps you, please:
- ⭐ Star this repository
- 🐦 Share it on social media
- 📝 Write about your experience
- 🗣️ Tell your team

---

## 🙏 Acknowledgments

Inspired by:
- [K6](https://k6.io/) - Performance testing excellence
- [Bruno](https://www.usebruno.com/) - Git-friendly API testing
- [OpenAPI](https://www.openapis.org/) - API specification standard

Built with ❤️ for developers who value simplicity and speed.

---

**Status**: 🚧 Early Development - v0.1.0-alpha

**Current Features**: Basic GET request testing with OpenAPI validation

Star ⭐ this repo to follow our progress!
