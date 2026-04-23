# CLI Generation Plan

Generate a Cobra-based CLI per Temporal service that wraps the existing
generated client helpers. The CLI exposes semantic, per-field command-line
flags derived from the workflow/signal/query input proto messages — not a
JSON blob.

Generation is **opt-in at the service level** with **per-workflow opt-out**.
Connection setup (address, namespace, TLS, auth) is **not** owned by the
generator; the caller builds a `client.Client` and passes it into the
generated entry point.

---

## 1. Proto option surface

### 1.1 Extend `ServiceOptions`

File: `proto/temporal/v1/temporal.proto`

```proto
message ServiceOptions {
  string task_queue = 1;
  WorkflowOptions default_workflow_options = 2;
  ActivityOptions default_activity_options = 3;

  // NEW — Generate a Cobra CLI for this service.
  // Default false. When true, a NewCLI(client.Client) *cobra.Command
  // function is emitted alongside the client/worker code.
  optional bool generate_cli = 4;
}
```

### 1.2 Extend `WorkflowOptions`

```proto
message WorkflowOptions {
  // ... existing fields ...

  // NEW — When true, this workflow is excluded from the generated CLI.
  // Use for workflows whose inputs can't be reasonably expressed as flags
  // (extreme nesting, heavy dependence on Any/Struct, etc.).
  // Default false. Only meaningful when the enclosing service has
  // generate_cli = true.
  optional bool skip_cli = 8;
}
```

### 1.3 Regeneration

After editing the proto:
```
make gen-tprl && make gen
```

---

## 2. Model layer changes

All new code lives under `internal/model/`. The existing `service.go`,
`workflow.go`, `signal.go`, `query.go` get a new sibling: `flagplan.go`.

### 2.1 `FlagPlan` — the core data structure

A `FlagPlan` describes how to turn one proto message into a set of
cobra/pflag flags. One `FlagPlan` per workflow input, per signal input,
per query input, plus one ad-hoc plan for the schedule options surface.

```go
// Package: internal/model

type FlagPlan struct {
    // Root Go type for the message this plan represents (for codegen).
    GoTypeName   string
    GoTypePkg    string
    // Ordered, deterministic list of flags to register.
    Flags        []*FlagSpec
    // Oneof groups, each element names the flag names that belong to it.
    // Used to emit MarkFlagsMutuallyExclusive calls.
    OneofGroups  [][]string
}

type FlagSpec struct {
    // Fully-qualified flag name as it appears on the command line,
    // kebab-case, path-joined with '-'.
    //   e.g. Order.Shipping.Street  -> "shipping-street"
    CLIName       string

    // Help text. MUST be phrased so the user does NOT have to guess the
    // input shape. See §6.
    Help          string

    // Proto field path from the root message (snake_case segments).
    // Used at assembly time to set the field on the proto struct.
    //   e.g. ["shipping", "street"]
    ProtoPath     []string

    // Kind drives which pflag function to call and which post-processing
    // is needed.
    Kind          FlagKind

    // Only populated for Kind==Enum.
    EnumGoType    string
    EnumValues    []string   // valid UPPER_SNAKE values for help text

    // Only populated for Kind==JSONEscape. Indicates the Go type the
    // JSON will be unmarshaled into (so the generated code can call
    // protojson.Unmarshal into a typed value).
    JSONGoType    string

    // Oneof membership (empty string if not in a oneof).
    OneofName     string

    // Whether the field is proto3 `optional` / a WKT wrapper — drives
    // whether presence is checked via Flags().Changed() before assignment.
    Optional      bool
}

type FlagKind int

const (
    FlagKindString FlagKind = iota
    FlagKindStringSlice     // repeated string -> StringArrayVar
    FlagKindInt32
    FlagKindInt32Slice
    FlagKindInt64
    FlagKindInt64Slice
    FlagKindUint32
    FlagKindUint64
    FlagKindFloat32
    FlagKindFloat64
    FlagKindBool
    FlagKindBytes           // base64 via BytesBase64Var
    FlagKindDuration        // DurationVar
    FlagKindTimestamp       // custom Value, RFC3339
    FlagKindEnum            // custom Value, validated against EnumValues
    FlagKindFieldMask       // StringSliceVar of paths
    FlagKindStringToString  // map<string,string>
    FlagKindStringToInt     // map<string,int>
    FlagKindJSONEscape      // repeated message / map of message / Any / Struct / cyclic
)
```

### 2.2 Building the plan — recursion rules

`NewFlagPlan(msg *protogen.Message) (*FlagPlan, error)` walks the message
descriptor breadth-first. Rules:

1. **Scalar leaf** → one `FlagSpec` with the corresponding scalar `FlagKind`.
2. **Enum leaf** → `FlagKindEnum`, populate `EnumValues` from the enum
   descriptor.
3. **Well-known type leaf** — special-case by full name:
   - `google.protobuf.Timestamp` → `FlagKindTimestamp`
   - `google.protobuf.Duration` → `FlagKindDuration`
   - `google.protobuf.FieldMask` → `FlagKindFieldMask`
   - `google.protobuf.{String,Int32,Int64,UInt32,UInt64,Float,Double,Bool,Bytes}Value`
     → the underlying scalar kind with `Optional=true` (presence gives the
     wrapper semantics).
   - `google.protobuf.Empty` → skip (emit no flag).
   - `google.protobuf.{Any,Struct,Value}` → `FlagKindJSONEscape`.
4. **Nested message leaf** (non-WKT) → recurse, prefix all child
   `CLIName`s with `parent-`.
5. **`repeated` scalar/enum/Timestamp/Duration** → the slice variant of the
   corresponding `FlagKind` (e.g. `FlagKindStringSlice`). Repeatable flag.
6. **`repeated <Message>`** → `FlagKindJSONEscape`. Single `--<field>-json`
   flag that accepts a JSON array literal. Do NOT recurse into the element.
7. **`map<string, scalar>`** → `FlagKindStringToString` /
   `FlagKindStringToInt` etc. Use pflag's `StringToStringVar` family.
8. **`map<_, Message>` or non-string key** → `FlagKindJSONEscape`.
9. **Oneof member** → set `OneofName` on each member's `FlagSpec`; append
   the group to `OneofGroups`. Emit each branch as its normal kind (scalar
   or nested). pflag + `MarkFlagsMutuallyExclusive` enforces exclusivity.
10. **Recursive type detection** — maintain a `visited` set keyed by proto
    full name while recursing. If a descent revisits a name already on the
    current path, fall back to `FlagKindJSONEscape` on the field that
    would have started the cycle. Do not stop the whole generation.
11. **Deterministic ordering** — iterate fields in proto field-number
    order at every level.

### 2.3 Name derivation

- Proto field `snake_case_name` → flag segment `snake-case-name` (replace
  `_` with `-`).
- Joined with parent path using `-` separator.
- If a proto field has a `json_name` override that differs from the
  default, prefer it (kebab-cased) so users can shorten collisions
  without editing code.
- Flag names are validated for uniqueness within a single `FlagPlan`.
  Collision is a generator error (should be rare — path prefixes prevent
  most of them).

### 2.4 Attaching plans to model objects

- `Workflow` gains: `InputFlagPlan *FlagPlan`, `SkipCLI bool`.
- `Signal`   gains: `InputFlagPlan *FlagPlan`.
- `Query`    gains: `InputFlagPlan *FlagPlan`.
- `Service`  gains: `GenerateCLI bool`.

`NewWorkflow`/`NewSignal`/`NewQuery` call `NewFlagPlan` when the enclosing
service has `GenerateCLI=true` AND (for workflows) `SkipCLI=false`.

---

## 3. Generated output — shape

### 3.1 Entry point

```go
// NewCLI returns the root cobra command for interacting with
// <ServiceName> workflows. The caller is responsible for constructing
// and closing the temporal client.
func New<ServiceName>CLI(c client.Client) *cobra.Command { ... }
```

One `NewXxxCLI` per service. Users wire it into their own `main.go`:

```go
root := &cobra.Command{Use: "myctl"}
root.AddCommand(myservicev1.NewOrdersServiceCLI(tc))
root.Execute()
```

### 3.2 Command tree (per service)

```
<service>
├── workflow
│   ├── start <WorkflowName> [flags...]        # StartWorkflow, prints run/wf IDs
│   ├── execute <WorkflowName> [flags...]      # start + Get, prints result protojson
│   ├── signal <WorkflowName> <SignalName> --workflow-id <id> [flags...]
│   ├── query <WorkflowName> <QueryName> --workflow-id <id> [flags...]
│   ├── cancel --workflow-id <id> [--run-id <id>]
│   ├── terminate --workflow-id <id> [--run-id <id>] [--reason <s>]
│   └── describe --workflow-id <id> [--run-id <id>]
└── schedule
    ├── create <WorkflowName> --schedule-id <id> [schedule flags...] [wf input flags...]
    ├── pause   --schedule-id <id> [--note <s>]
    ├── unpause --schedule-id <id> [--note <s>]
    ├── delete  --schedule-id <id>
    └── describe --schedule-id <id>
```

- `<WorkflowName>` / `<SignalName>` / `<QueryName>` are **subcommand
  names** generated per workflow/signal/query, not positional args. That
  is, `workflow start` is itself a group with a subcommand per workflow.
  Shown flattened above for readability; actual tree:
  `service workflow start order-submit --customer-id ...`.

- Signal/query commands are nested under the workflow they belong to:
  `service workflow signal order-submit cancel-order --workflow-id ...`.
  This mirrors how the Go client wires signals to a specific workflow.

### 3.3 Per-workflow subcommand generation

For each workflow with a plan, emit:

- `new<Workflow>StartCmd(c client.Client) *cobra.Command`
- `new<Workflow>ExecuteCmd(c client.Client) *cobra.Command`
- `new<Workflow>SignalCmd<Signal>(c client.Client) *cobra.Command` (one per signal listed in `WorkflowOptions.signals`)
- `new<Workflow>QueryCmd<Query>(c client.Client) *cobra.Command`  (one per query listed in `WorkflowOptions.queries`)
- `new<Workflow>ScheduleCreateCmd(c client.Client) *cobra.Command`

Each command's `RunE`:
1. Allocates a zero-value request proto.
2. For each `FlagSpec`, reads the flag value and sets the field via the
   `ProtoPath`. Presence-gated fields skip assignment when
   `cmd.Flags().Changed(name) == false`.
3. Calls the corresponding already-generated client method (e.g.
   `clientImpl.StartOrderSubmit(ctx, req, opts...)`).
4. On success, prints result using `protojson.MarshalOptions{Multiline: true, Indent: "  "}`.

### 3.4 Shared infrastructure

Emit once per file:

- `type timestampValue struct{ ts *timestamppb.Timestamp }` implementing `pflag.Value` (String, Set, Type), accepting RFC3339 + RFC3339Nano.
- `type enum<EnumGoName>Value struct{ v *<EnumGoName> }` per enum referenced by any plan. `Set` accepts case-insensitive names, errors with the full set in the message.
- `type jsonValue[T proto.Message] struct{ raw string; target T }` with deferred `protojson.Unmarshal` called in `RunE` (so the flag can be left empty without failing).
- Helper `applyPresence(cmd *cobra.Command, name string, apply func())` that runs `apply` only when `cmd.Flags().Changed(name)`.

---

## 4. Flag-name conventions

- All kebab-case.
- Nested message fields prefix with parent path: `shipping-address-street`.
- Schedule root flags prefix with `schedule-`: `--schedule-id`, `--schedule-cron`.
- Workflow-input flags on `schedule create` use the workflow input's
  flag names directly (no `input-` prefix). Schedule and input namespaces
  are disjoint because schedule-specific flags all start with `schedule-`.
- `--workflow-id` and `--run-id` are reserved for identifying a running
  workflow. Input messages with a `workflow_id` field would collide; in
  that case the input flag becomes `--input-workflow-id`. Handle this at
  plan build time: if a workflow command would register `--workflow-id`
  from input AND needs it for targeting, rename the input's to
  `input-workflow-id` and document it in the flag's help text.

---

## 5. Schedule flags (scoped)

Emit on `schedule create` only:

| Flag                   | pflag type         | Meaning |
|------------------------|--------------------|---------|
| `--schedule-id`        | StringVar (required) | Schedule identifier. |
| `--schedule-cron`      | StringArrayVar     | Cron expression. Repeatable to add multiple cron rules. |
| `--schedule-interval`  | DurationSliceVar   | Interval spec. Repeatable. e.g. `5m`, `1h`. |
| `--schedule-start-at`  | custom Timestamp   | RFC3339. Don't start firing before this. |
| `--schedule-end-at`    | custom Timestamp   | RFC3339. Stop firing after this. |
| `--schedule-timezone`  | StringVar          | IANA name, e.g. `Europe/Paris`. |
| `--schedule-jitter`    | DurationVar        | Max random offset applied to each firing. |
| `--schedule-paused`    | BoolVar            | Start in paused state. |
| `--schedule-note`      | StringVar          | Free-text note attached to the schedule state. |
| `--schedule-overlap`   | custom Enum        | Overlap policy: `skip` \| `buffer-one` \| `buffer-all` \| `cancel-other` \| `terminate-other` \| `allow-all`. |
| `--schedule-catchup-window` | DurationVar   | Max catch-up window. |
| `--schedule-task-queue`| StringVar          | Override task queue for this schedule's runs. Defaults to service task queue. |

**Calendar-style spec flags are explicitly not generated.** Users who need
calendar specs build the `client.ScheduleOptions` themselves in Go.

For `pause` / `unpause`: `--schedule-id` (required), `--note` (optional).
For `delete` / `describe`: `--schedule-id` only.

---

## 6. Help text — MUST be unambiguous about input shape

This is non-negotiable. Every generated flag's help string states the
expected input format explicitly. No "oh is that a comma-list or JSON?"
moments. Rules:

### 6.1 Templates per kind

| Kind | Help-string template |
|------|----------------------|
| string | `<proto-comment>` (trimmed) |
| int/float/bool | `<proto-comment>` |
| bytes | `<proto-comment> (base64-encoded)` |
| duration | `<proto-comment> (Go duration, e.g. "5m", "1h30m")` |
| timestamp | `<proto-comment> (RFC3339 timestamp, e.g. "2025-01-02T15:04:05Z")` |
| enum | `<proto-comment> (one of: FOO, BAR, BAZ; case-insensitive)` |
| fieldmask | `<proto-comment> (comma-separated field paths, e.g. "name,address.city")` |
| string slice | `<proto-comment> (repeatable flag: pass --<name> once per element)` |
| int32/int64 slice | `<proto-comment> (repeatable flag: pass --<name> once per element)` |
| string->string map | `<proto-comment> (repeatable flag in key=value form, e.g. --<name> foo=bar --<name> baz=qux)` |
| string->int map | `<proto-comment> (repeatable flag in key=value form with integer values)` |
| JSON escape (repeated message) | `<proto-comment> (JSON array of <MessageType>, e.g. '[{"field":"value"}]')` |
| JSON escape (map of message)   | `<proto-comment> (JSON object mapping <KeyType> to <MessageType>, e.g. '{"a":{"field":"value"}}')` |
| JSON escape (Any/Struct/Value) | `<proto-comment> (JSON value)` |
| JSON escape (cyclic subtree)   | `<proto-comment> (JSON value of type <MessageType>; this field is self-referencing and cannot be expressed via individual flags)` |

- `<proto-comment>` is the trailing / leading comment from the field,
  trimmed, with newlines replaced by spaces. If absent, fall back to
  `<Kebab-case field name>`.
- The parenthetical format hint is appended with a leading space.
- For `FlagKindJSONEscape` flags the example JSON MUST name the target
  type so the user knows what shape to fill in.

### 6.2 `MarkFlagsMutuallyExclusive` groups

Also mentioned in help via cobra's built-in annotations + an extra line
in the command's `Long` description listing the oneof groups:

```
Mutually exclusive flag groups:
  payment:  --payment-card-number, --payment-bank-account
```

### 6.3 Required flags

`--workflow-id`, `--schedule-id`, positional `<WorkflowName>` command
names — all marked via `cobra.ExactArgs` or `cmd.MarkFlagRequired`.
Proto3 has no "required" semantics for input message fields; don't
invent one. If a server-side validation rejects missing fields, the
error is surfaced verbatim.

---

## 7. Template layer changes

New template file: `internal/renderer/templates/cli.tmpl`.

`service.tmpl` gets a new gate near the bottom:

```gotmpl
{{- if .GenerateCLI }}
{{ template "cli.tmpl" . }}
{{- end }}
```

`cli.tmpl` sections (in order):

1. Imports (cobra, pflag, protojson, timestamppb, durationpb, etc).
2. `NewXxxCLI` top-level function.
3. Per-workflow command constructors (start, execute, signal*, query*, schedule-create).
4. Schedule pause/unpause/delete/describe (one each, not per-workflow — they act on schedule IDs).
5. Cancel / terminate / describe (workflow-id targeted).
6. Shared `pflag.Value` implementations (timestamp, JSON, enum per type).
7. Small helpers (`setString`, `setInt32`, `setTimestamp`, etc., to walk `ProtoPath` and assign fields).

### 7.1 Template helpers (add to `internal/tmpl/funcs.go`)

- `kebab(s string) string` — snake_case/camelCase → kebab.
- `pflagFuncName(kind FlagKind) string` — maps `FlagKind` to the pflag
  binder name (`StringVar`, `StringArrayVar`, ...).
- `goFieldAssignPath(path []string, rootVar string) string` — emits
  `rootVar.Shipping.GetAddress().Street` style access, with `Get*` for
  nullable message intermediates, allocating as needed.
- `enumValueTypeName(enumGoName string) string` — e.g. `OrderStatus` → `enumOrderStatusValue`.
- `jsonExampleFor(kind FlagKind, msgName string) string` — produces the
  illustrative JSON snippet used in help text.

### 7.2 `FuncMap` registration

`FuncMap(gf)` in `internal/tmpl/funcs.go` gets the new helpers added. The
existing `commentBlock` helper is reused for help-text derivation.

---

## 8. Generator wiring (`main.go` / protogen entry)

- Read the new `generate_cli` service option and set `Service.GenerateCLI`.
- Read `skip_cli` on each workflow's `WorkflowOptions` → `Workflow.SkipCLI`.
- Build `FlagPlan`s lazily: only when `Service.GenerateCLI && !Workflow.SkipCLI`.
- All CLI code emits into the **same generated file** as the existing
  client/worker code. No new output file.

---

## 9. Unsupported-input handling

At `NewFlagPlan` time, the walker may decide a whole field is unflaggable
and fall back to `FlagKindJSONEscape`. It never refuses to produce a
plan — every opted-in workflow gets a plan, with JSON-escape islands
where needed. Per §1.2, workflows whose inputs are hopeless can be
opted out individually via `skip_cli`.

One hard error remains: if the workflow's input message itself is the
cyclic type (root of the recursion), degrade to a single top-level
`--input-json` flag and log a warning via `protogen.Plugin.Error` that
degrades to stderr (so users notice).

---

## 10. Testing strategy

### 10.1 Unit tests

- `internal/model/flagplan_test.go`
  - Every `FlagKind` produced for synthetic descriptors.
  - WKT special cases.
  - Recursive type detection → JSON escape on the cyclic field.
  - Oneof group assembly.
  - Name collision detection.
  - `json_name` override honored.
  - `--workflow-id` / `--run-id` collision renaming.

- `internal/renderer/renderer_test.go`
  - Snapshot test: a service with one workflow using a rich input
    message renders deterministically. Update snapshot only via
    `go test ./... -update`.

- `internal/tmpl/funcs_test.go`
  - `kebab`, `pflagFuncName`, `goFieldAssignPath` unit tests.

### 10.2 Example-driven integration

- Extend `example/proto/*.proto` with a new service that opts into the
  CLI and has representative fields: scalars, an enum, a nested message,
  a timestamp, a duration, `repeated string`, `repeated <Message>`, a
  scalar map, a oneof. Regenerate with `make gen`.
- Add `example/cli/main.go` that wires `NewXxxCLI` into a Cobra root.
- `verify-examples` target updated to `go build ./example/cli/...`.

### 10.3 Execution tests

- `gen/example/v1/cli_test.go` exercises commands via
  `cobra.Command.SetArgs` + a fake `client.Client` (interface-based
  stub), asserts that a given flag invocation produces the expected
  request proto and calls the expected client method.

---

## 11. Dependency notes

- New Go module deps on the generator module:
  - none — cobra and pflag are **runtime** deps of generated code.
- Add to generated-code imports (already pulled by consumers):
  - `github.com/spf13/cobra`
  - `github.com/spf13/pflag`
  - `google.golang.org/protobuf/encoding/protojson`
  - `google.golang.org/protobuf/types/known/{timestamppb,durationpb,fieldmaskpb}`
- Schedule command uses `go.temporal.io/sdk/client` types already in use
  by the existing generated code; no new Temporal SDK imports.

---

## 12. Documentation

- `README.md` gets a new top-level section "Generating a CLI" with:
  - Minimal proto example enabling `generate_cli`.
  - Example wiring in `main.go`.
  - A table of proto-type → flag-type mappings (same content as §6.1).
  - The `skip_cli` opt-out.
- Generated `documentation.tmpl` markdown output gets a new per-service
  "CLI" section listing the command tree and flags. Gated on
  `.GenerateCLI`.

---

## 13. Subagent task breakdown

The following tasks are independent enough to be parallelized after Task 1 lands. Subsequent tasks depend on earlier ones as noted.

**Every subagent MUST write and maintain a progress file** — see §14. Without it a crash or user interruption wastes all the work the agent has done so far.

### Task 1 — Proto option + model scaffolding (sequential, foundation)
**Files:** `proto/temporal/v1/temporal.proto`, `internal/model/service.go`, `internal/model/workflow.go`, `internal/model/options.go`, regenerated `gen/temporal/v1/temporal.pb.go`.
**Scope:**
- Add `generate_cli` to `ServiceOptions`, `skip_cli` to `WorkflowOptions`.
- Run `make gen-tprl`.
- Thread `GenerateCLI`, `SkipCLI` through the model.
- Unit tests covering option plumbing.
**Done when:** `make test-unit` passes with new option-plumbing tests; no generator output changes yet.

### Task 2 — FlagPlan builder (depends on Task 1)
**Files:** `internal/model/flagplan.go`, `internal/model/flagplan_test.go`.
**Scope:** Implement the full walker per §2. All `FlagKind`s. All WKT special cases. Recursion detection. Oneof grouping. Name collision/reservation handling. Exhaustive unit tests on synthetic descriptors.
**Done when:** `go test ./internal/model/... -race` green.

### Task 3 — Template helpers (parallel with Task 2)
**Files:** `internal/tmpl/funcs.go`, `internal/tmpl/funcs_test.go`.
**Scope:** `kebab`, `pflagFuncName`, `goFieldAssignPath`, `enumValueTypeName`, `jsonExampleFor`. Wire into `FuncMap`.
**Done when:** `go test ./internal/tmpl/... -race` green.

### Task 4 — CLI template (depends on Tasks 2 and 3)
**Files:** `internal/renderer/templates/cli.tmpl`, `internal/renderer/templates/service.tmpl`, `internal/renderer/renderer_test.go`.
**Scope:** The full `cli.tmpl` per §7. Gate in `service.tmpl`. Snapshot test.
**Done when:** `go test ./internal/renderer/... -race` green; snapshot checked in.

### Task 5 — Example service + generated CLI verification (depends on Task 4)
**Files:** `example/proto/**/*.proto`, regenerated `gen/example/v1/*.go`, `example/cli/main.go`, `gen/example/v1/cli_test.go`, `Makefile`.
**Scope:** Add the opt-in example service covering the representative input shapes from §10.2. Add `example/cli/main.go`. Extend `verify-examples`. Add execution tests for generated commands.
**Done when:** `make` (full pipeline) green.

### Task 6 — Documentation (parallel with Task 5)
**Files:** `README.md`, `internal/renderer/templates/documentation.tmpl`.
**Scope:** README section on CLI generation; per-service CLI section in generated markdown docs.
**Done when:** README renders, `make gen` emits docs with CLI section for opted-in services.

### Task 7 — Schedule command polish (depends on Task 5)
**Files:** `internal/renderer/templates/cli.tmpl`, tests under `gen/example/v1/`.
**Scope:** Overlap-policy enum mapping, jitter + catchup wiring, pause/unpause/delete/describe end-to-end tests against a fake client.
**Done when:** schedule subcommands covered by integration tests.

---

## 14. Progress tracking & recovery

All subagent work is incremental and resumable. If a subagent is interrupted, killed, or the user `/clear`s mid-task, the next agent picks up from the progress file — no re-discovery, no duplicated work.

### 14.1 Location and naming

- Directory: `.plan-progress/` at the repo root.
- **This directory MUST be gitignored** — add `.plan-progress/` to `.gitignore` before the first subagent runs. Progress files are local coordination, not artifacts to commit.
- One file per task: `task-01-proto-scaffolding.md`, `task-02-flagplan-builder.md`, etc. (zero-padded, numeric, kebab slug of the task title).
- A single roll-up file: `.plan-progress/STATUS.md` with one line per task showing current state. Agents update their own line; they do NOT rewrite other tasks' lines.

### 14.2 Progress file format (per task)

Each task file uses this exact skeleton so any later agent can parse it at a glance:

```markdown
# Task N — <Title>

- **Status:** not-started | in-progress | blocked | done
- **Started:** <ISO-8601 timestamp, or "—" if not-started>
- **Last updated:** <ISO-8601 timestamp>
- **Owner agent run:** <short description of the current invocation, or "—">

## Scope recap
<1–3 lines pulled from §13, so the file is self-contained>

## Checklist
- [ ] <sub-step 1>
- [ ] <sub-step 2>
- ...

## Decisions & deviations
<running log of any judgment calls made that deviate from PLAN.md, with a one-line reason. Append-only.>

## Blockers
<current blockers if Status=blocked, else "none". Clear when unblocked.>

## Artifacts touched
<running list of files created/edited with one-line purpose, append-only>

## Notes for the next agent
<any context that would save the next agent time: "the X interface already exists at Y", "don't re-run make gen yet, it's broken because...", etc.>
```

### 14.3 Subagent protocol

Every subagent invoked to implement a task MUST:

1. **On start:**
   - Read `PLAN.md` (the source of truth for what to do) AND its task progress file (the source of truth for what's already done).
   - If the progress file doesn't exist, create it from the skeleton with `Status: in-progress` and a populated checklist derived from the task's scope.
   - If it exists with `Status: done`, stop immediately and report back — the task is finished.
   - If it exists with `Status: in-progress` or `blocked`, resume from the first unchecked checklist item.
   - Update `Last updated` and `Owner agent run`.
   - Update the task's line in `.plan-progress/STATUS.md`.

2. **During work:**
   - Check off checklist items **as they are completed**, not in a batch at the end.
   - Append any non-trivial decision to the "Decisions & deviations" section, with a one-line reason.
   - Append each edited/created file to "Artifacts touched".
   - If blocked (missing info, failing test can't be diagnosed in-scope, etc.), set `Status: blocked`, write a concrete description under "Blockers", and stop.

3. **On completion:**
   - Run the task's "Done when" check from §13. If it passes, set `Status: done`, update `Last updated`, clear the "Blockers" section, and write a short handoff in "Notes for the next agent" (e.g. "run `make gen` before Task N — I didn't, to keep this PR scoped").
   - Update `.plan-progress/STATUS.md`.

4. **Hard rules:**
   - Never delete or rewrite history in the progress file. Edits are append-only for "Decisions & deviations", "Artifacts touched", and "Notes for the next agent". Checklist items flip from `[ ]` to `[x]` in place.
   - Never edit another task's progress file.
   - Never commit `.plan-progress/` content. If `git status` shows the directory as tracked, investigate — it should be ignored.
   - If the progress file and reality disagree (e.g. checklist says a file was created but it doesn't exist), trust reality. Correct the file with a "Decisions & deviations" entry explaining the drift.

### 14.4 Orchestrator protocol (parent agent spinning subagents)

Before spinning a subagent for Task N:
- Read `.plan-progress/STATUS.md` to know current state of all tasks.
- If Task N has unmet dependencies per §13, refuse to spin and report which prerequisite is incomplete.
- Pass the subagent: the task number, the path to its progress file, and an instruction to follow §14.3.

After a subagent returns:
- Read its progress file to determine actual state (subagent reports can be optimistic). Trust the file.
- Update the orchestrator's plan accordingly.

### 14.5 Bootstrap

Before Task 1 begins:
- Ensure `.gitignore` contains `.plan-progress/`.
- Create `.plan-progress/STATUS.md` with one line per task, all `not-started`.
- Do NOT pre-create per-task files. Each subagent creates its own on first run.

---

## 15. Out of scope (deliberately)

- Generating a full `main.go` binary. Users wire `NewXxxCLI` into their
  own root command and own the client lifecycle.
- Connection-config flags (`--address`, `--namespace`, TLS). Belongs in
  the caller's own root command or a wrapper.
- `workflow list` / generic search. No equivalent in the existing
  generated client — would need standalone implementation, not CLI
  generation territory.
- Activity direct invocation from CLI. Activities are invoked by
  workflows, not by end users.
- Structured calendar schedule specs. Users who need them write Go.
- Interactive prompts for missing fields. All values come from flags.
