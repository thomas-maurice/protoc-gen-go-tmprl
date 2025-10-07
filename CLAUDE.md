# Basics
- Don't read generated code, it's a waste of tokens

# Bash commands
- make : Runs ALL tests (unit + k8sclient integration) and builds everything (DEFAULT - use this)
- make test : Runs ALL tests (unit tests + k8sclient integration test)
- make test-unit : Runs only unit tests with race detection and coverage
- make test-k8sclient : Runs k8sclient integration test with Kind cluster
- make bench : Runs all benchmarks (conversion and Lua access performance)
- make bench-update : Runs benchmarks AND updates benchmarks/README.md with results
- make build : Builds all binaries (stubgen, example)
- make clean : Removes built binaries

# GitHub Actions local testing (act)
- make act-check : Check if act is installed
- make act-list : List all GitHub Actions workflows and jobs
- make act-test-unit : Run unit tests job locally (fast, recommended for pre-commit checks)
- make act-lint : Run lint job locally
- make act-build : Run build job locally
- make act-test : Run ALL test jobs locally (includes K8s integration, slow)

# Answer style
- Be brief until instructed otherwise. Don't assume the user is dumb.
- Don't use emojis, they are bullshit.

# Code style
- Use the standard go comment style when commenting go code: "<funcName>: <description>[optionally \n<example usage>]"
- Comment the code for any non trivial operations
- Write unit tests whenever it is possible for any new function added
- Don't use emojis, they are bullshit

# Workflow
- **CRITICAL**: Always run `make` before considering a task complete - ALL tests must pass (unit + integration)
- **CRITICAL**: The whole repo's tests must pass otherwise a commit isn't an option
- **CRITICAL**: If pre-commit hooks are available, run them before any commit
- **RECOMMENDED**: Use `make act-test-unit` to verify GitHub Actions will pass before pushing
- `make` runs: unit tests, k8sclient integration test with Kind cluster, and builds all binaries
- Prefer running single tests during development (`go test ./pkg/...`), but ALWAYS run `make` before finishing
- When adding new features, add tests in the same package (pkg/glua/, pkg/modules/*, etc.)
- Use `act` to test GitHub Actions locally and catch CI failures early

# Test Requirements
- Unit tests: Must pass with race detection (`go test -race -cover ./pkg/...`)
- Integration tests: k8sclient example must pass with Kind cluster (`make test-k8sclient`)
- Kind and kubectl must be installed for integration tests
- All tests are run automatically by `make` or `make test`

# Benchmarks
- **CRITICAL**: Before ANY commit, ALWAYS run `make bench-update` to update benchmark results
- **CRITICAL**: When making performance-related changes, verify benchmarks show expected improvements
- Run benchmarks: `make bench` (view only) or `make bench-update` (update README)
- Benchmarks track conversion performance and Lua field access patterns
- The pre-commit hook does NOT check this automatically - you must run it manually
