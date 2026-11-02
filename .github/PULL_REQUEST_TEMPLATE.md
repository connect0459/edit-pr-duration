<!-- # PULL_REQUEST_TEMPLATE -->

<!-- Remove unnecessary sections to keep the review focused -->

## Related Links

- Issues
  - <!-- <https://github.com/connect0459/edit-pr-duration/issues/xxx> -->
- PRs
  - <!-- <https://github.com/connect0459/edit-pr-duration/pull/xxx> -->

## [Required] Overview

- Describe the problem being solved, its background, and what changes when this PR is merged.
- Links to `docs/ARCHITECTURE.md`, ADRs under `docs/adrs/`, or other references are welcome.

```txt
It is difficult to review without knowing the specifications and background.
```

## Scope of Change

- [ ] `internal/domain` package(s)
- [ ] `internal/infrastructure` package(s)
- [ ] `internal/application` package(s)
- [ ] `pkg/spinner`
- [ ] `main.go` (CLI entrypoint)
- [ ] Tooling / CI
- [ ] Documentation (`docs/`, README)

## Breaking Changes

- [ ] No breaking changes
- [ ] Breaking changes (describe below)

<!--
If this changes the `config.json` schema or the CLI flags, describe what breaks and why the breakage is justified, and update README.md and config.example.json to match.
-->

## Deferred Items and TODOs

- Items intentionally deferred and the reasons why.

```txt
If you deferred something due to time constraints, document it here.
Reviewers cannot tell whether something was intentionally skipped or overlooked
without this information.
```

## Test Items

- Describe the tests added, following Red/Green TDD (which test was written first, and what it confirmed failed before the implementation existed).
- Note per-layer coverage if it changed meaningfully (see the test strategy in `docs/ARCHITECTURE.md`).
- Confirm `gofmt -l .`, `go vet ./...`, `go build ./...`, and `go test ./...` all pass with no regressions.

## [Required] Quality Checklist

**Please check all items before merging.**

- [ ] **CI Workflow Execution**: All checks passed on the [CI workflow](../actions/workflows/ci.yml) for this PR.
- [ ] **Code Comments**: Limited to Godoc and non-obvious WHY/WHY-NOT explanations, per this project's comment policy.
- [ ] **Reference Docs**: `docs/ARCHITECTURE.md` and ADRs updated if the design changed.

> **Important**: This checklist ensures quality. Please verify all items before requesting review.
