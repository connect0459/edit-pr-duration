# Contributing

## Prerequisites

- [Go](https://go.dev/dl/) — version pinned in `go.mod`
- [gh CLI](https://cli.github.com/) — authenticated (`gh auth login`); the tool shells out to `gh` to read and edit PRs
- [pre-commit](https://pre-commit.com/) — hook runner

## Setup

```sh
git clone https://github.com/connect0459/edit-pr-duration
cd edit-pr-duration
pre-commit install
```

`pre-commit install` installs the pre-commit hooks.

To run all hooks manually:

```sh
pre-commit run --all-files
```

To try the tool locally, copy `config.example.json` to `config.json`, edit it, and run a dry run first:

```sh
go run . --config config.json --dry-run --verbose
```

## Development workflow

This mirrors the CI pipeline (`.github/workflows/ci.yml`), in order:

| Command | Purpose |
| :--- | :--- |
| `gofmt -l .` | Check formatting (empty output = clean) |
| `go vet ./...` | Static analysis |
| `go build ./...` | Build |
| `go test ./...` | Run all tests |

The pre-commit hooks enforce `gofmt`, `go vet`, and Markdown lint checks on every commit.

## Testing guidelines

This project follows **Red → Green → Refactor** (Detroit-school TDD):

- Write a failing test first, then implement.
- Use real objects; mocks are only permitted at external boundaries (file system, the `gh` CLI, network). Domain-level collaboration is tested with the in-memory `GitHubRepository` in `internal/infrastructure/memory`.
- A test names the **exported** unit it exercises as its prefix (`Test<ExportedUnit>_<BehaviorDescription>`) — an unexported unit's correctness is demonstrated through the exported entry point that calls it, not by testing the unexported unit directly, unless a comment documents why the exported entry point cannot reach that behavior.
- Error messages describe a concrete operation or state, not the name of the function/method that produced them.

Coverage targets are agreed on before starting an implementation; no fixed numeric floor is enforced in CI. Use real fixtures (`t.TempDir()`) for file I/O.

## Commit format

```text
<type>(<scope>): <subject>

<body>

<footer>
```

**Types**: `feat`, `fix`, `docs`, `style`, `refactor`, `tidy`, `test`, `chore`, `ci`, `perf`

**Scope**: package name when the change targets a specific package (e.g. `application`, `ghcli`, `spinner`); omit for project-wide changes. Use a domain-axis type (`docs`, `style`, `test`, `chore`, `ci`, `tidy`) as `type` when the change is fully contained within that domain, rather than as `scope` on an impact-axis type.

**Subject**: imperative mood ("add", "fix", "remove"), 72 characters max, no trailing period.

**Body** (optional): wrap at 72 characters; explain **why**, not what — the diff already shows what changed.

**Footer** (optional): `BREAKING CHANGE: <description>`, or `Closes <https://github.com/owner/repo/issues/123>` to link issues.

## Pull request process

1. Fork the repository and create a branch: `feat/xxx`, `fix/xxx`, `docs/xxx`.
2. Follow the Red → Green → Refactor cycle; keep one concern per commit.
3. Run `go test ./...` and `pre-commit run --all-files`, and ensure both pass.
4. Open a pull request using the repository's PR template — CI (`gofmt`, `go vet`, `go build`, `go test`) runs automatically.

## Code style

- Onion architecture: `internal/domain` defines abstract types (ports, value objects); `internal/infrastructure` implements them; `internal/application` orchestrates across layers; `main.go` at the repository root is the CLI entrypoint. No layer depends directly on a concrete type from a layer it doesn't own.
- Rich domain objects: pair data and logic in the same type; prefer immutability.
- Avoid code comments unless the **why** is genuinely non-obvious — let tests document behavior. Exported functions/methods carry a one-line Godoc.
- Record significant design decisions as ADRs under `docs/adrs/` (start from `adr-000-template.md`).
- All identifiers, test names, error messages, and documentation must be in **English** (see `AGENTS.md`).
