# Basics
- Don't read generated code, it's a waste of tokens

# Bash commands
- make : Runs ALL tests and builds everything (DEFAULT - use this)
- make test : Runs ALL tests
- make test-unit : Runs only unit tests with race detection and coverage
- make build : Builds the plugin binary
- make gen : Regenerates example code
- make verify-examples : Builds and verifies example code compiles
- make clean : Removes built binaries

# GitHub Actions local testing (act)
- make act-check : Check if act is installed
- make act-list : List all GitHub Actions workflows and jobs
- make act-test-unit : Run unit tests job locally (fast, recommended for pre-commit checks)
- make act-lint : Run lint job locally
- make act-build : Run build job locally
- make act-test : Run ALL test jobs locally

# Answer style
- Be brief until instructed otherwise. Don't assume the user is dumb.
- Don't use emojis, they are bullshit.

# Code style
- Use the standard go comment style when commenting go code: "<funcName>: <description>[optionally \n<example usage>]"
- Comment the code for any non trivial operations
- Write unit tests whenever it is possible for any new function added
- Don't use emojis, they are bullshit

# Workflow
- **CRITICAL**: Always run `make` before considering a task complete - ALL tests must pass
- **CRITICAL**: The whole repo's tests must pass otherwise a commit isn't an option
- **CRITICAL**: If pre-commit hooks are available, run them before any commit
- **RECOMMENDED**: Use `make act-test-unit` to verify GitHub Actions will pass before pushing
- `make` runs: unit tests with race detection, regenerates examples, verifies examples build, and builds the plugin binary
- Prefer running single tests during development (`go test ./internal/...`), but ALWAYS run `make` before finishing
- When adding new features, add tests in the same package (internal/model/, internal/tmpl/, etc.)
- Use `act` to test GitHub Actions locally and catch CI failures early

# Test Requirements
- Unit tests: Must pass with race detection (`go test -race -cover ./internal/...`)
- Example verification: Generated examples must compile (`make verify-examples`)
- All tests are run automatically by `make` or `make test`
