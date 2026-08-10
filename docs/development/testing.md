# Testing

`cani` has two kinds of tests.

**Go unit tests** live beside the code in `*_test.go` and exercise packages in-process. Run them with
`make utest`.

**ShellSpec tests** live under `spec/` and exercise the *compiled binary*. They assert on the things a Go
unit test cannot reach: exit codes, what lands on stdout versus stderr, flag parsing, help text, and the
state of the datastore after a command runs. That is the CLI contract, and it is what users actually depend
on.

Most contributors only need `make test-fast` (unit + functional). See
[CONTRIBUTING.md](https://github.com/Cray-HPE/cani/blob/main/.github/CONTRIBUTING.md) for the test tiers and
prerequisites.

## Getting ShellSpec

```shell
make spec-setup
```

This clones a pinned ShellSpec into `.tools/` and symlinks it into `.tools/bin`. No `sudo`, nothing outside
the repo. The `ftest`, `itest`, and `etest` targets add `.tools/bin` to `PATH` themselves, so you do not need
to. If ShellSpec is already on your `PATH`, the target leaves it alone.

`make spec-clean` removes it again.

## Anatomy of a spec

`.shellspec` loads `spec/spec_helper.sh` for every run, which is where the shared environment and helpers
come from. A spec is a tree of `Describe` blocks containing `It` examples:

```shell
Describe 'cani alpha update location'
  Before 'setup_crud_env'

  Describe 'valid name'
    It 'updates a location by name'
      When call bin/cani alpha update location test-site --description "updated" --config "$CANI_CONF"
      The status should equal 0
      The stderr should include 'Updated location'
    End
  End
End
```

The pieces:

| Directive | Purpose |
| --- | --- |
| `Describe` / `Context` | Group examples. Nest freely; the names concatenate in the output |
| `It` | One example. Exactly one `When` per example |
| `When call <cmd>` | Run `<cmd>` in the current shell and capture status, stdout, stderr |
| `When run <cmd>` | Same, but in a subshell. Use when the command may `exit` |
| `The status should equal 0` | Assert the exit code |
| `The stdout should include '...'` | Assert on output. Also `should equal`, `should match pattern`, `should start with` |
| `Before` / `After` | Run a helper before/after every example in the block |
| `BeforeCall` / `BeforeAll` | Run just before the `When`, or once per block |
| `Skip if '<reason>' <fn>` | Skip the block when `<fn>` returns 0 |

`cani` builds to `bin/cani`, and the helper puts that directory on `PATH`, so specs invoke `bin/cani`
directly.

!!! warning "`Skip if` takes a command, not a shell expression"

    `Skip if 'reason' ! curl ...` is **silently ignored** — `!` is a shell keyword, so ShellSpec never sees a
    condition and the block runs anyway. Wrap the negation in a function instead:

    ```shell
    nautobot_unreachable() {
      ! curl -sf -H "Authorization: Token ${NAUTOBOT_TOKEN}" "${NAUTOBOT_URL}/status/" >/dev/null 2>&1
    }
    Skip if 'Nautobot is not reachable' nautobot_unreachable
    ```

    A bare `[ ... ]` test is fine, because `[` is a real command.

## Environment and helpers

`spec_helper_precheck()` exports these for every spec:

| Variable | Value | Notes |
| --- | --- | --- |
| `FIXTURES` | `spec/testdata/fixtures` | Root for all fixture files |
| `CANI_DIR` | `/tmp/.cani` | Shared scratch dir, kept off your real config |
| `CANI_CONF` / `CANI_DS` / `CANI_LOG` | under `CANI_DIR` | Config, datastore, log |
| `NAUTOBOT_URL` / `NAUTOBOT_TOKEN` | `localhost:8081`, a fake token | Must match `cani_0.6.x.yml` |
| `SKIP_EXTERNAL_TESTS` | `1` unless `RUN_EXTERNAL_TESTS=1` | External-service tests are opt-in |

Environment builders — each wipes `CANI_DIR` and lays down a known state:

| Helper | Gives you |
| --- | --- |
| `setup_test_env` | Empty config, no datastore |
| `setup_populated_env` | The `test-rack` inventory |
| `setup_crud_env` | test-site, test-rack, test-device, test-module, test-cable, an interface |
| `setup_connections_env` | CRUD inventory for connection tests |
| `setup_orphan_env` | Orphaned devices and racks |
| `setup_migration_env <cfg>` | A legacy config fixture, e.g. `cani_0.1.x.yml` |
| `setup_datastore_migration_env <ds>` | A legacy datastore only, so `cani` must create the config |
| `setup_nautobot_env` | Config pointing at the local Nautobot |
| `teardown_test_env` | Removes `CANI_DIR` |

Also available: `remove_config`, `remove_datastore`, `remove_log`, `fixture` (compare a value to a fixture
file), `match_colored_text` / `match_rich_text` (match through ANSI escapes), and for Nautobot,
`wait_for_nautobot`, `nb_create`, `nb_list`, `nb_list_field`.

Because every builder wipes `/tmp/.cani`, a spec that dies midway can leave stale state behind. `rm -rf
/tmp/.cani` resets it.

## Functional or integration?

The deciding question is **does this test need a live service?**

- **No** → `spec/functional/`. Runs against local fixtures. `make ftest` needs nothing but ShellSpec.
- **Yes** → `spec/integration/`. Needs CSM on `:8443` or Nautobot on `:8081` via `make sim-up`.

Most changes want both. A functional spec proves the command parses its flags and prints what it should; only
an integration spec proves it put the right data in the provider. Integration coverage is expected when you
add a command or flag that touches a provider, add a provider, or modify an existing one — including its API
client, field mapping, or translation logic. The same applies to shared code every provider routes through,
such as the datastore, the inventory schema, or CRUD plumbing.

For a provider, the round trip is the thing worth asserting: import → modify → export → re-import, with the
last step proving idempotency.

Assert on the provider's API from inside the spec itself. `spec/spec_helper.sh` provides `nb_list`,
`nb_list_field`, and `nb_create` for Nautobot; for anything else, wrap a `curl` call in a helper function the
way `spec/integration/nautobot_import_export_spec.sh` does.

!!! warning "Tavern assertions are deprecated"

    Some CSM specs still pair with a [tavern](https://tavern.readthedocs.io) assertion file, matched by
    filename: `spec/integration/foo_spec.sh` is followed by `spec/integration/tavern/test_foo.tavern.yml`
    when that file exists, and `spec/support/bin/cani_integrate.sh` wires the two together.

    They are kept for backwards compatibility only — **do not write new ones.** Every one of them targets
    CSM, their endpoints are pinned to the legacy simulator on `localhost:8080`, and they are the reason
    `make itest` needs a Python virtualenv at all.

Guard anything that talks to a live service so the suite stays green without it:

```shell
Describe 'INTEGRATION: Nautobot round-trip'
  Skip if 'SKIP_EXTERNAL_TESTS is set' [ "${SKIP_EXTERNAL_TESTS:-0}" = "1" ]
  Skip if 'Nautobot is not reachable' nautobot_unreachable
```

## Running a subset

```shell
shellspec spec/functional/add_spec.sh          # one file
shellspec spec/functional/add_spec.sh:42       # the example at line 42
shellspec spec/functional --focus              # only blocks marked fDescribe/fIt
shellspec spec/functional -f documentation     # readable output instead of TAP
shellspec spec/functional/add_spec.sh -x       # trace every command (debugging)
```

Run `make bin` first — specs test the binary in `bin/`, not your working tree.

## Reading TAP output

`make ftest` reports TAP, which CI parses:

```
ok 12 - cani alpha add rack --help exits 0
not ok 13 - cani alpha add rack adds a rack
not ok 54 - is a Todo # TODO Not yet implemented
ok 3 - INTEGRATION: Nautobot round-trip imports # SKIP Nautobot is not reachable
```

`# TODO` and `# SKIP` do not fail the suite. A bare `not ok` does.

## Placeholders

To reserve a test for behaviour that is not implemented yet, write an example with no `When`:

```shell
It 'is a Todo'
End
```

