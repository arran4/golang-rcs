# Agent Instructions

## Testing
*   Use `go:embed` from `testdata/` for all multiline sample data in tests.
*   New features should include tests covering both positive and negative cases.
*   When testing CLI flags, prefer integration tests that verify behavior through the command execution path.

## Coding Standards
*   Use standard Go formatting (`go fmt`).
*   Ensure all new files have appropriate package declarations.
*   Avoid adding binary artifacts to the repository.
