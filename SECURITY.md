# Security Policy

## Supported Versions

Only the latest release is supported with security fixes. As this project is pre-1.0, no long-term support window is guaranteed for older releases.

## Reporting a Vulnerability

**Please do not open a public GitHub issue for security vulnerabilities.**

Use GitHub's [private vulnerability reporting][private-report] feature, or email <connect0459@gmail.com>, to disclose issues confidentially. You can expect an acknowledgement within 7 days and a status update within 30 days.

**What to include:**

- A minimal, reproducible example (a sanitized `config.json`, a sample PR body, or a CLI invocation) that triggers the issue.
- The version (commit) of `edit-pr-duration` and OS/architecture.
- A description of the impact.

[private-report]: https://github.com/connect0459/edit-pr-duration/security/advisories/new

## Scope

`edit-pr-duration` reads pull request metadata through the authenticated `gh` CLI, computes working-hours durations, and rewrites PR bodies on GitHub. The following classes of issues are in scope:

- **Unintended PR modification** — any input (a crafted PR body, a `config.json`, or a command-line argument) that causes the tool to edit a PR outside the configured `repositories.targets`, `period`, or `author`, or to edit a PR while `--dry-run` is set.
- **Command injection via `gh` invocation** — any configuration value or fetched PR content that is interpreted as an additional `gh` argument or shell command instead of being passed as data.
- **Resource exhaustion or catastrophic regular expressions** — a `replacement_patterns` entry or PR body that causes unbounded memory growth or unbounded CPU time.
- **Panics/crashes on malformed input** — any malformed config file or adversarial GitHub API response that causes a panic instead of a handled error.

The following are **out of scope**:

- Vulnerabilities in the `gh` CLI itself or GitHub's own API — report those to their respective maintainers.
- Risks that arise from running the tool with a `config.json` the operator did not review; the config file is trusted input from the local operator.
- Issues that require the attacker to already have write access to the target repositories' PRs beyond what the authenticated `gh` credential grants.

## Disclosure Policy

Once a fix is released, a GitHub Security Advisory will be published with full details. The typical timeline from report to public disclosure is 30 days, though this may be extended by mutual agreement when a fix requires significant changes.
