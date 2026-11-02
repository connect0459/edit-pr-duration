# edit-pr-duration

[![CI](https://github.com/connect0459/edit-pr-duration/actions/workflows/ci.yml/badge.svg)](https://github.com/connect0459/edit-pr-duration/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](https://github.com/connect0459/edit-pr-duration/blob/main/LICENSE)
[![GitHub CLI](https://img.shields.io/badge/gh-cli-blue.svg)](https://cli.github.com)

A Go CLI that calculates the working time of GitHub pull requests and writes it into the PR body, replacing a placeholder such as "Actual time spent: xx hours".

## Motivation

The time a PR actually took is easy to promise and easy to forget: the placeholder is filled in by hand after the merge, if at all. `edit-pr-duration` fills it in from the PR's own timestamps, so the recorded duration comes from the PR itself rather than from memory.

- **Working time, not wall-clock time.** Only weekday working hours count, with configurable holidays, so a PR left open over a weekend is not reported as 48 hours of work.
- **Many repositories in one run.** Target repositories, period, and author are all set in a single config file.
- **Safe to try.** `--dry-run` shows what would change without touching any PR.

## Installation

Prerequisites:

- [Go](https://go.dev/dl/) (version pinned in `go.mod`)
- [GitHub CLI](https://cli.github.com/) (`gh`) installed and authenticated (`gh auth login`)
- Access permission to the target repositories

To build from source:

```sh
git clone https://github.com/connect0459/edit-pr-duration.git
cd edit-pr-duration
go build
```

## Usage

### 1. Prepare the config file

```sh
cp config.example.json config.json
```

Edit `config.json` to set the target repositories, period, working hours, holidays, and replacement patterns (see [Config File](#config-file)).

### 2. Run

```sh
# Check what would change (dry-run mode)
./edit-pr-duration --dry-run

# Actually update PRs
./edit-pr-duration

# Specify a custom config file
./edit-pr-duration --config /path/to/config.json

# Print verbose output
./edit-pr-duration --verbose
```

`--dry-run` and `--verbose` are command-line flags only; they cannot be set in `config.json`.

## Config File

The config file (`config.json`) accepts the following settings:

### Author

Only PRs created by this GitHub user are processed.

```json
{
  "author": "your-github-username"
}
```

### Target Repositories

```json
{
  "repositories": {
    "targets": [
      "organization/repository1",
      "organization/repository2"
    ]
  }
}
```

### Target Period

```json
{
  "period": {
    "start_date": "2025-10-01T00:00:00Z",
    "end_date": "2025-12-19T23:59:59Z"
  }
}
```

### Working Hours

```json
{
  "work_hours": {
    "start_hour": 9,
    "start_minute": 30,
    "end_hour": 18,
    "end_minute": 30
  }
}
```

### Holidays

```json
{
  "holidays": [
    {
      "dates": [
        "2025-10-14",
        "2025-11-04"
      ]
    }
  ]
}
```

### Replacement Patterns

Regular expression patterns that replace a placeholder in the PR body with the actual working time. Patterns are tried in order, and the first one that matches is used. If no pattern matches, the PR is not updated (`replacement_patterns` is required).

- `pattern`: The regular expression to match (capture group 1 can preserve the leading text)
- `replacement`: The replacement template (`{hours}` expands to the formatted duration string)
- `hours_format`: The duration notation (`"ja"` or `"en"`)

In the example below, the first pattern matches a Japanese placeholder (`実際にかかった時間` = "actual time spent", `XX 時間` = "XX hours") and writes the duration in the Japanese notation; the second matches an English placeholder and writes it in the English notation.

```json
{
  "replacement_patterns": [
    {
      "pattern": "(実際にかかった時間\\s*[:：]?\\s*\\r?\\n?\\s*[-*]?\\s*)(?:約?\\s*)?(?:XX|xx)\\s*時間",
      "replacement": "${1}{hours}",
      "hours_format": "ja"
    },
    {
      "pattern": "(Actual time spent\\s*\\r?\\n\\s+[-*]\\s*)(?:about\\s*)?(?:XX|xx)\\s*hours",
      "replacement": "${1}{hours}",
      "hours_format": "en"
    }
  ]
}
```

## Documentation

- [Architecture](https://github.com/connect0459/edit-pr-duration/blob/main/docs/ARCHITECTURE.md) — onion architecture layers, project structure, and test strategy.
- [ADRs](https://github.com/connect0459/edit-pr-duration/tree/main/docs/adrs) — recorded design decisions.

## Contributing

See [CONTRIBUTING.md](https://github.com/connect0459/edit-pr-duration/blob/main/CONTRIBUTING.md).

## License

[MIT](https://github.com/connect0459/edit-pr-duration/blob/main/LICENSE)
