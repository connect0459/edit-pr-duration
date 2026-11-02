# Architecture Overview - edit-pr-duration

## Project Overview

A CLI tool that automatically replaces the working-time placeholder in a GitHub PR body (e.g. "Actual time spent: xx hours") with the actual working time between PR creation and merge/close.

## 1. Project Structure

```text
edit-pr-duration/
├── main.go                          # CLI entry point, dependency wiring
├── config.example.json              # Sample config file
├── go.mod                           # Go module definition
├── README.md                        # Project description
├── CONTRIBUTING.md                  # Development workflow and conventions
├── .gitignore                       # Git ignore settings
├── docs/                            # Architecture overview and ADRs
├── pkg/
│   └── spinner/                     # Progress spinner for the CLI
└── internal/
    ├── domain/                      # Domain layer (business logic)
    │   ├── entities/                # Entities (have an identity)
    │   │   ├── config.go           # Config entity
    │   │   ├── prinfo.go           # PR info entity
    │   │   └── prinfo_test.go
    │   ├── valueobjects/            # Value objects (no identity)
    │   │   ├── period.go           # Target period
    │   │   ├── workhours.go        # Working hours
    │   │   └── options.go          # Run options
    │   ├── services/                # Domain services
    │   │   └── calculator.go       # Working-time calculation logic
    │   └── repositories/            # Repository abstract types (interfaces)
    │       ├── config_repository.go
    │       └── github_repository.go
    ├── application/                 # Application layer (use cases)
    │   ├── service.go              # PRDurationService
    │   └── service_test.go         # Integration tests
    └── infrastructure/              # Infrastructure layer (external systems)
        ├── json/                    # JSON config loading
        │   ├── config_repository.go
        │   └── config_repository_test.go
        ├── ghcli/                   # GitHub CLI implementation
        │   └── github_repository.go
        └── memory/                  # In-memory implementation for tests
            └── github_repository.go
```

### Layer Responsibilities

| Layer | Responsibility | Dependency direction |
| --- | --- | --- |
| **Domain** | Business logic, domain model | No external dependencies (abstract types only) |
| **Application** | Use cases, coordination of domain objects | Depends on the Domain layer only |
| **Infrastructure** | External system integration, persistence | Implements the Domain layer's abstract types |
| **Main** | Dependency wiring, CLI argument handling | Depends on all layers |

## 2. System Diagram

```text
┌─────────────┐
│    User     │
└──────┬──────┘
       │ runs command
       ▼
┌─────────────────────────────────────┐
│         main.go (CLI)               │
│  - Loads config                     │
│  - Wires dependencies               │
│  - Prints results                   │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│   Application Layer                 │
│   ┌─────────────────────────────┐   │
│   │  PRDurationService          │   │
│   │  - Run(): bulk PR update    │   │
│   └─────────────────────────────┘   │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│   Domain Layer                      │
│   ┌──────────┐  ┌──────────────┐    │
│   │ Entities │  │ ValueObjects │    │
│   │ - Config │  │ - Period     │    │
│   │ - PRInfo │  │ - WorkHours  │    │
│   └──────────┘  │ - Options    │    │
│                 └──────────────┘    │
│   ┌──────────────────────────────┐  │
│   │ Services                     │  │
│   │ - Calculator (working time)  │  │
│   └──────────────────────────────┘  │
│   ┌──────────────────────────────┐  │
│   │ Repositories (abstract)      │  │
│   │ - ConfigRepository           │  │
│   │ - GitHubRepository           │  │
│   └──────────────────────────────┘  │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│   Infrastructure Layer              │
│   ┌────────────┐  ┌──────────────┐  │
│   │ JSON       │  │ GitHub CLI   │  │
│   │ Config     │  │ Repository   │  │
│   │ Reader     │  │ (gh wrapper) │  │
│   └──────┬─────┘  └──────┬───────┘  │
│          │                │          │
└──────────┼────────────────┼──────────┘
           │                │
           ▼                ▼
    ┌──────────┐    ┌──────────────┐
    │ config   │    │  GitHub API  │
    │ .json    │    │  (gh CLI)    │
    └──────────┘    └──────────────┘
```

## 3. Core Components

### 3.1 Domain Layer

#### Entities

| Component | Responsibility | Identity |
| --- | --- | --- |
| **Config** | Holds the whole application configuration | File path (implicit) |
| **PRInfo** | PR information and placeholder replacement logic | Repository name + PR number |

**Design principles:**

- Always-Valid Domain Model (immutable entities)
- Rich Domain Objects (contain business logic)

#### Value Objects

| Component | Responsibility |
| --- | --- |
| **Period** | Target period (StartDate, EndDate) |
| **WorkHours** | Working hours (start/end time) |
| **Options** | Run options (DryRun, Verbose) |

#### Services (Domain Services)

| Component | Responsibility |
| --- | --- |
| **Calculator** | Working-time calculation (counts weekday working hours only) |

### 3.2 Application Layer

| Component | Responsibility |
| --- | --- |
| **PRDurationService** | Implements the bulk PR update use case |

**Main processing flow:**

1. Get the target repositories and period from the config
2. Fetch the matching PR list through the GitHub API
3. Calculate the working time of each PR (using Calculator)
4. Replace the placeholder (PRInfo.UpdatedBody())
5. Update the PR through the GitHub API (honors dry-run mode)

### 3.3 Infrastructure Layer

| Component | Technology | Responsibility |
| --- | --- | --- |
| **json.ConfigRepository** | encoding/json | Loads the JSON config file |
| **ghcli.GitHubRepository** | os/exec | GitHub CLI (gh) wrapper |
| **memory.GitHubRepository** | in-memory | Test double (Detroit school) |

## 4. Data Stores

| Kind | Description | Format |
| --- | --- | --- |
| **Config file** | config.json | JSON (standard library) |
| **GitHub PRs** | Via the GitHub API | REST/GraphQL API (gh CLI) |

### Config File Structure (config.json)

The `hours_format: "ja"` pattern below matches a Japanese placeholder (`実際にかかった時間` = "actual time spent", `xx 時間` = "xx hours") and writes the duration in the Japanese notation.

```json
{
  "author": "your-github-username",
  "repositories": {"targets": ["org/repo1", "org/repo2"]},
  "period": {
    "start_date": "2025-10-01T00:00:00Z",
    "end_date": "2025-12-31T23:59:59Z"
  },
  "work_hours": {
    "start_hour": 9,
    "start_minute": 30,
    "end_hour": 18,
    "end_minute": 30
  },
  "holidays": [{"dates": ["2025-10-14", "2025-11-04"]}],
  "replacement_patterns": [
    {"pattern": "(実際にかかった時間\\s*)xx\\s*時間", "replacement": "${1}{hours}", "hours_format": "ja"}
  ]
}
```

## 5. External Integrations / APIs

| Integration | Purpose | Authentication |
| --- | --- | --- |
| **GitHub CLI (gh)** | Fetch/update PR information | gh auth login |

### GitHub CLI Operations

```bash
# List PRs (filtered by --author when an author is configured)
gh pr list --repo org/repo --state all --limit 1000 --json number,createdAt

# Fetch PR details (GraphQL)
gh api graphql --input -

# Update a PR
gh pr edit 123 --repo org/repo --body "new body"
```

## 6. Deployment & Infrastructure

### Build

```bash
go build -o edit-pr-duration
```

### Run

```bash
# Dry-run mode (does not update)
./edit-pr-duration --config config.json --dry-run

# Actually update
./edit-pr-duration --config config.json

# Verbose output
./edit-pr-duration --config config.json --verbose
```

### Prerequisites

- Go (version pinned in `go.mod`)
- GitHub CLI (gh) installed
- Authenticated with `gh auth login`

## 7. Security Considerations

| Item | Measure |
| --- | --- |
| **GitHub authentication** | Uses gh CLI's authentication (token management is left to gh CLI) |
| **Config file** | config.json is excluded via .gitignore |
| **Dry-run mode** | Lets you preview changes before running for real |

## 8. Development & Testing Environment

### Local Setup

```bash
# No external dependencies (standard library only)
go mod download

# Build
go build

# Run tests
go test ./...

# Check coverage
go test -cover ./...
```

### Test Strategy

| Layer | Test type | Coverage target |
| --- | --- | --- |
| **Domain** | Unit tests (Test Object Pattern) | 90% or more |
| **Application** | Integration tests (using the memory implementation) | 80% or more |
| **Infrastructure** | Integration tests (real files) | 70% or more |

**Test approach:**

- **TDD (Test-Driven Development)** - Red → Green → Refactor
- **Detroit school** - Mocks only at external boundaries (GitHub API)
- **Test Object Pattern** - Test data is managed as structs
- **Living Documentation** - Test names express the specification

### Code Quality

```bash
# Format
gofmt -l .

# Static analysis
go vet ./...
```

## 9. Future Considerations / Roadmap

### Technical Debt

1. **Raise test coverage**
   - services/calculator_test.go (not yet written)
   - entities/config_test.go (not yet written)

### Feature Ideas

1. **Parallel processing**
   - Currently: PR updates are processed sequentially
   - Idea: parallelize with goroutines

2. **Better logging**
   - Currently: standard output only
   - Idea: structured logs (JSON format)

## 10. Project Identification

| Item | Detail |
| --- | --- |
| **Project name** | edit-pr-duration |
| **Repository** | <https://github.com/connect0459/edit-pr-duration> |
| **Language/Framework** | Go (standard library only) |
| **Architecture** | Onion Architecture |
| **Last updated** | 2025-12-19 |
| **Maintainer** | [connect0459](https://github.com/connect0459) |

## 11. Glossary / Abbreviations

| Term | Description |
| --- | --- |
| **PR** | Pull Request (GitHub) |
| **Entity** | A domain object with an identity (immutable under the Always-Valid Domain Model) |
| **Value Object** | A domain object without an identity (immutable) |
| **ADR** | Architecture Decision Record |
| **TDD** | Test-Driven Development |
| **DDD** | Domain-Driven Design |
| **Dry-run** | A mode that only verifies behavior without actually updating anything |
| **Test Object Pattern** | A pattern that manages test data as structs |
| **Detroit-school TDD** | A TDD style that minimizes mocks and emphasizes collaboration between real objects |
| **gh CLI** | GitHub's official command-line tool |
| **Working time** | Counts weekday working hours only (9:30-18:30) |

---

**About this document:**
This architecture document is a living document for understanding the technical big picture of the project. When the design changes, update it together with the corresponding ADR (`docs/adrs/`).
