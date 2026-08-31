# Contributing

## Quick start

```shell
git clone https://github.com/Cray-HPE/cani.git && cd cani
make hooks        # enable local commit-message validation
make spec-setup   # install shellspec into .tools/ (no sudo)
make doctor       # check prerequisites, print a fix for anything missing
make test-fast    # unit + functional tests; no Docker, no Python
```

`make doctor` is the source of truth for prerequisites. It prints a table of what is present, what is
missing, and the exact command that fixes each gap. Run it with `TIER=int` to also require everything the
integration tests need.

## Prerequisites

| Tool | Needed for | Install |
| --- | --- | --- |
| Go (see `go.mod`) | everything | <https://go.dev/dl/> |
| shellspec >= 0.28.1 | functional + integration tests | `make spec-setup` |
| GNU sed (`gsed`) | functional tests on macOS only | `brew install gnu-sed` |
| Docker + Compose v2 | integration tests | <https://docs.docker.com/get-docker/> |
| Python 3.12 + `virtualenv` | integration tests (legacy CSM tavern assertions) | `pip3 install virtualenv`, then `make venv` |
| `jq`, `curl`, `openssl` | integration tests and cert generation | your package manager |

The CSM simulator images come from `artifactory.algol60.net`, which is HPE-internal. Run
`docker login artifactory.algol60.net` before `make csm-up`. The Nautobot simulator uses public images and
needs no login.

## Test tiers

Pick the smallest tier that covers your change. Each tier builds on the one above it.

| Tier | Command | Requires | Proves |
| --- | --- | --- | --- |
| Unit | `make utest` | Go | Package-level logic, in-process |
| Functional | `make ftest` | shellspec | The CLI contract: exit codes, stdout/stderr, flags, help text |
| Fast | `make test-fast` | Go + shellspec | Unit + functional. The daily loop. **No Docker, no Python.** |
| Integration | `make itest` | simulators + `venv` | End-to-end workflows against live CSM/Nautobot APIs |
| Everything | `make test` | all of the above | What CI approximates |
| Edge | `make etest` | `venv` | Boundary cases. Excluded from `make test` (slow, EOL CSM) |

`make test` stands the simulators up for you. To run the integration tier by hand:

```shell
make sim-up      # CSM on :8443, Nautobot on :8081
make itest
make sim-down
```

### External-service tests

Tests that need a live Nautobot or CSM API are **skipped by default**, so a fresh clone is green without any
simulator running. Opt in once the simulators are up:

```shell
RUN_EXTERNAL_TESTS=1 make itest
```

Setting `SKIP_EXTERNAL_TESTS` explicitly overrides both defaults, which is how CI pins the behaviour. Tests
also skip themselves individually when the service they need is unreachable, so opting in on a machine where
only one simulator is running is safe.

## Writing tests

Go unit tests live beside the code in `*_test.go`. Everything that exercises the compiled binary is a
shellspec file under `spec/`. See [docs/development/testing.md](../docs/development/testing.md) for the
shellspec primer: spec anatomy, the shared helpers, fixtures, and how to run a single example.

Where a new test belongs:

- `spec/functional/` — the command runs against local fixtures only. No network, no services.
- `spec/integration/` — the command needs a live CSM or Nautobot API.

If you add a new command, add a `spec/functional/<name>_spec.sh` that follows the shape of its neighbours.
Each flag and each distinct output should be accounted for. If you need a placeholder for behaviour that is
not ready yet, add a Todo-style test:

```shell
It 'is a Todo'
End
```

It reports as `# TODO` while still letting the suite pass:

```
not ok 54 - is a Todo # TODO Not yet implemented
```

Add fixtures under `testdata/fixtures/` as needed.

### When integration coverage is required

Functional tests prove a command parses its flags and prints the right thing. They cannot prove it wrote the
right thing to a provider. Add or extend a `spec/integration/` spec whenever you:

- **add a command or a flag** that reads from or writes to a provider
- **add a provider**, which needs round-trip coverage: import → modify → export → re-import for idempotency
- **modify an existing provider**, including its API client, field mapping, or translation logic

A change to shared code that every provider goes through — the datastore, the inventory schema, CRUD
plumbing — needs integration coverage for at least one provider, because that is the only place the
end-to-end contract is exercised.

Assert against the provider's API from inside the spec, using `curl` and the helpers in
`spec/spec_helper.sh`. **Do not add tavern files.** They are deprecated, cover CSM only, and are kept purely
for backwards compatibility. Guard anything that talks to a live service so the suite stays green when the
simulators are down; see the testing primer for the `Skip if` pattern.

## Before you open a pull request

```shell
make fmt vet
make lint          # golangci-lint; install once with make install-lint
make lint-size     # 300-line budget for non-test Go files
make test-fast
```

Conventions enforced in CI:

- **Conventional commits.** Commit subjects and the PR title are validated. `make hooks` catches this locally
  before you push.
- **File size.** Non-test Go files stay under 300 lines. `tools/file_size_baseline.txt` is a ratchet: files
  already over budget may shrink but not grow.
- **Layering.** `cmd/` contains no provider-specific logic. Provider code belongs in `pkg/provider/<name>/`
  and hooks into commands through the interfaces in `internal/provider/`. See [AGENTS.md](../AGENTS.md).
- **License headers.** New files carry the MIT header used throughout the repo.
- **Shell scripts** must pass `shellcheck`.

## Attribution trailers

Whether a change was AI-assisted cannot be detected, only declared. `make hooks` installs a local check
that validates whatever you do declare, so the trailer block stays machine readable:

```
feat(export): add rack filtering

Assisted-by: GitHub Copilot
Co-authored-by: Ada Lovelace <ada@example.com>
Signed-off-by: Grace Hopper <grace@example.com>
```

- `Assisted-by:` names the tool that helped. A tool is not an identity, so it carries no mail address.
- `Co-authored-by:` credits a person or account and needs `Name <user@host>`. Use an address tied to a real
  GitHub account, otherwise the co-authorship does not register.
- `Signed-off-by:` needs `Name <user@host>` and must be a human. An AI cannot certify a contribution; the
  person sending it owns every line, including the generated ones.

Trailers must follow a blank line, and only the trailers you write are checked — none are required.

## Troubleshooting

| Symptom | Cause | Fix |
| --- | --- | --- |
| `shellspec: command not found` | not installed | `make spec-setup` (the make targets put `.tools/bin` on `PATH` themselves) |
| `sed: unknown option` on macOS | BSD sed lacks GNU address ranges | `brew install gnu-sed` |
| `unauthorized` pulling `cray-sls` or `cray-smd` | not logged in to the HPE registry | `docker login artifactory.algol60.net` |
| Port 8081 or 8443 already in use | a simulator is still running | `make sim-down` |
| Integration tests fail on stale inventory | leftover state in the shared test dir | `rm -rf /tmp/.cani` |
| Nautobot tests skip unexpectedly | external tests are opt-in | `make sim-up`, then `RUN_EXTERNAL_TESTS=1 make itest` |
| A commit is rejected locally | conventional-commit hook | match `type(scope): subject`, e.g. `fix(export): handle empty rack` |
| A trailer is rejected locally | attribution-trailer hook | see [Attribution trailers](#attribution-trailers); tools use `Assisted-by:`, humans sign off |
