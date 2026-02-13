# Agent Instructions

## Testing
*   Use `go:embed` from `testdata/` for all multiline sample data in tests.
*   Use `.txtar` files in `testdata/txtar/` for parser and serializer round-trip tests.
    *   Each `.txtar` file should contain an `input.rcs` file and an `expected.json` file.
    *   These tests verify that parsing `input.rcs` results in the structure defined in `expected.json`, and that `File.String()` matches the original input (or a normalized version if specified).
*   New features should include tests covering both positive and negative cases.
*   When testing CLI flags, prefer integration tests that verify behavior through the command execution path.

## Coding Standards
*   Use standard Go formatting (`go fmt`).
*   Ensure all new files have appropriate package declarations.
*   Avoid adding binary artifacts to the repository.

## Documentation
*   Whenever subcommands are updated, update the documentation such as `readme.md`.
