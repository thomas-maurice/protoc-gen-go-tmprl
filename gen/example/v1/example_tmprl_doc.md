<a id="top"></a>
# Services
<a id="service_example_v1_DieRoll"></a>
## example.v1.DieRoll
Service DieRoll is an example implementation of a service
 It doesn't do much

 But it is there, chilling.

 This documentation will be generated along the code
 ```golang
 package main

 import "fmt"

 func main() {
     fmt.Println("You can also put markdown in there, how cool is that ?")
 }
 ```

### Table of contents

   * [example.v1.DieRoll default settings](#svcoptions_example_v1_DieRoll)
 * Workflows
   * [example.v1.DieRoll.ParentWorkflow](#method_example_v1_DieRoll_ParentWorkflow)
   * [example.v1.DieRoll.ChildWorkflow](#method_example_v1_DieRoll_ChildWorkflow)
   * [example.v1.DieRoll.ThrowDies](#method_example_v1_DieRoll_ThrowDies)
   * [example.v1.DieRoll.ThrowUntilValue](#method_example_v1_DieRoll_ThrowUntilValue)
 * Activities
   * [example.v1.DieRoll.ThrowDie](#method_example_v1_DieRoll_ThrowDie)
   * [example.v1.DieRoll.Ping](#method_example_v1_DieRoll_Ping)
 * Signals
   * [example.v1.DieRoll.Continue](#method_example_v1_DieRoll_Continue)
 * Queries
   * [example.v1.DieRoll.GetThrowsStatus](#method_example_v1_DieRoll_GetThrowsStatus)

<a id="svcoptions_example_v1_DieRoll"></a>
### Service options
| Option | Value |
| --- | --- |
| Default task queue | `service-task-queue` |

### Default workflow options
| Option | Value |
| --- | --- |
| Workflow execution timeout | 0s |
| Workflow run timeout | 0s |

### Workflows
<a id="method_example_v1_DieRoll_ParentWorkflow"></a>
#### example.v1.DieRoll.ParentWorkflow
Parent workflow that calls the Child workflow -- to test workflow ID generations mainly

Input: [google.protobuf.Empty](#message_google_protobuf_Empty)

Output: [example.v1.ParentWorkflowReply](#message_example_v1_ParentWorkflowReply)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `example.v1.DieRoll.ParentWorkflow` |
| Workflow execution timeout | 24h0m0s |
| Workflow run timeout | 2h0m0s |

Signals:
 * [example.v1.DieRoll.Continue](#method_example_v1_DieRoll_Continue)
<a id="method_example_v1_DieRoll_ChildWorkflow"></a>
#### example.v1.DieRoll.ChildWorkflow


Input: [google.protobuf.Empty](#message_google_protobuf_Empty)

Output: [google.protobuf.Empty](#message_google_protobuf_Empty)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `example.v1.DieRoll.ChildWorkflow` |
| Workflow execution timeout | 24h0m0s |
| Workflow run timeout | 2h0m0s |
<a id="method_example_v1_DieRoll_ThrowDies"></a>
#### example.v1.DieRoll.ThrowDies
Throws dies a few times and return the result

Input: [example.v1.ThrowDiesRequest](#message_example_v1_ThrowDiesRequest)

Output: [example.v1.ThrowDiesResponse](#message_example_v1_ThrowDiesResponse)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `example.v1.DieRoll.ThrowDies` |
| Workflow execution timeout | 24h0m0s |
| Workflow run timeout | 2h0m0s |

Signals:
 * [example.v1.DieRoll.Continue](#method_example_v1_DieRoll_Continue)
<a id="method_example_v1_DieRoll_ThrowUntilValue"></a>
#### example.v1.DieRoll.ThrowUntilValue


Input: [example.v1.ThrowUntilValueRequest](#message_example_v1_ThrowUntilValueRequest)

Output: [google.protobuf.Empty](#message_google_protobuf_Empty)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `example.v1.DieRoll.ThrowUntilValue` |
| Workflow execution timeout | 24h0m0s |
| Workflow run timeout | 2h0m0s |

Queries:
 * [example.v1.DieRoll.GetThrowsStatus](#method_example_v1_DieRoll_GetThrowsStatus)

### Activities
<a id="method_example_v1_DieRoll_ThrowDie"></a>
#### example.v1.DieRoll.ThrowDie
Throws a d6 and returns the result

Input: [google.protobuf.Empty](#message_google_protobuf_Empty)

Output: [example.v1.ThrowDieResponse](#message_example_v1_ThrowDieResponse)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `example.v1.DieRoll.ThrowDie` |
| Schedule to close timeout | 2m0s |
| Schedule to start timeout | 30s |
| Start to close timeout | 2m0s |

Retry policy:

| Option | Value |
| --- | --- |
| Initial interval | 1s |
| Backoff coefficient | 1.500000 |
| Maximum attempts | 10 |
| Maximum interval | 10s |
| Non retryable error types | `FATAL`, `NOT_FOUND` |
<a id="method_example_v1_DieRoll_Ping"></a>
#### example.v1.DieRoll.Ping
Just a simple ping
 Takes no parameters
 returns nothing

Input: [google.protobuf.Empty](#message_google_protobuf_Empty)

Output: [google.protobuf.Empty](#message_google_protobuf_Empty)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `ping.Ping` |
| Schedule to close timeout | 24h0m0s |
| Heartbeat timeout | 1m0s |

### Queries
<a id="method_example_v1_DieRoll_GetThrowsStatus"></a>
#### example.v1.DieRoll.GetThrowsStatus
Query the state of the workflow

Input: [google.protobuf.Empty](#message_google_protobuf_Empty)

Output: [example.v1.ThrowStatusResponse](#message_example_v1_ThrowStatusResponse)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `example.v1.DieRoll.GetThrowsStatus` |

### Signals
<a id="method_example_v1_DieRoll_Continue"></a>
#### example.v1.DieRoll.Continue
Instruct the workflow to proceed

Input: [example.v1.ContinueSignalRequest](#message_example_v1_ContinueSignalRequest)

Output: [google.protobuf.Empty](#message_google_protobuf_Empty)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `example.v1.DieRoll.Continue` |

<a id="cli_example_v1_DieRoll"></a>
## CLI

This service is opted into CLI generation (`generate_cli = true`). The
generated entry point `NewDieRollCLI(client.Client) *cobra.Command`
returns a root subcommand that the caller plugs under its own Cobra root.

The root command is `die-roll` and groups the following
subcommands:

```
die-roll
|-- workflow
|   |-- start     <workflow>
|   |-- execute   <workflow>
|   |-- signal    <workflow> <signal>
|   |-- query     <workflow> <query>
|   |-- cancel
|   |-- terminate
|   `-- describe
`-- schedule
    |-- create    <workflow>
    |-- pause
    |-- unpause
    |-- delete
    `-- describe
```

### Workflow commands

#### `die-roll workflow start parent-workflow` / `execute parent-workflow` / `schedule create parent-workflow`

Input flags (derived from `google.protobuf.Empty`):

_(no input flags)_

##### `die-roll workflow signal parent-workflow continue`

Targeting flags: `--workflow-id` (required), `--run-id` (optional).

Signal payload flags (derived from `example.v1.ContinueSignalRequest`):

| Flag | Kind | Description |
| --- | --- | --- |
| `--continue` | Bool | continue |

#### `die-roll workflow start child-workflow` / `execute child-workflow` / `schedule create child-workflow`

Input flags (derived from `google.protobuf.Empty`):

_(no input flags)_

#### `die-roll workflow start throw-dies` / `execute throw-dies` / `schedule create throw-dies`

Input flags (derived from `example.v1.ThrowDiesRequest`):

| Flag | Kind | Description |
| --- | --- | --- |
| `--results` | Int32 | Result array |
| `--loop` | Bool | Loop ? |
| `--result-status` | String | A deprecated field |

##### `die-roll workflow signal throw-dies continue`

Targeting flags: `--workflow-id` (required), `--run-id` (optional).

Signal payload flags (derived from `example.v1.ContinueSignalRequest`):

| Flag | Kind | Description |
| --- | --- | --- |
| `--continue` | Bool | continue |

#### `die-roll workflow start throw-until-value` / `execute throw-until-value` / `schedule create throw-until-value`

Input flags (derived from `example.v1.ThrowUntilValueRequest`):

| Flag | Kind | Description |
| --- | --- | --- |
| `--value` | Int32 | Target value |

##### `die-roll workflow query throw-until-value get-throws-status`

Targeting flags: `--workflow-id` (required), `--run-id` (optional).

### Cross-workflow commands

These act on a running workflow by ID. They do not take workflow-input flags.

| Command | Required flags | Optional flags |
| --- | --- | --- |
| `die-roll workflow cancel` | `--workflow-id` | `--run-id` |
| `die-roll workflow terminate` | `--workflow-id` | `--run-id`, `--reason` |
| `die-roll workflow describe` | `--workflow-id` | `--run-id` |

### Schedule commands

`schedule create <workflow>` accepts the workflow's input flags (see above)
plus the schedule flags below.

| Flag | Kind | Description |
| --- | --- | --- |
| `--schedule-id` | String | Schedule identifier (required). |
| `--schedule-cron` | StringSlice | Cron expression. Repeatable for multiple cron rules. |
| `--schedule-interval` | DurationSlice | Interval spec. Repeatable, e.g. `5m`, `1h`. |
| `--schedule-start-at` | Timestamp | RFC3339. Do not fire before this time. |
| `--schedule-end-at` | Timestamp | RFC3339. Stop firing after this time. |
| `--schedule-timezone` | String | IANA timezone name, e.g. `Europe/Paris`. |
| `--schedule-jitter` | Duration | Max random offset applied to each firing. |
| `--schedule-paused` | Bool | Start the schedule in paused state. |
| `--schedule-note` | String | Free-text note attached to the schedule state. |
| `--schedule-overlap` | Enum | `skip` \| `buffer-one` \| `buffer-all` \| `cancel-other` \| `terminate-other` \| `allow-all`. |
| `--schedule-catchup-window` | Duration | Max catch-up window. |
| `--schedule-task-queue` | String | Override task queue for this schedule's runs. |

Schedule lifecycle commands (keyed by schedule ID, no workflow input):

| Command | Required flags | Optional flags |
| --- | --- | --- |
| `die-roll schedule pause` | `--schedule-id` | `--note` |
| `die-roll schedule unpause` | `--schedule-id` | `--note` |
| `die-roll schedule delete` | `--schedule-id` | _(none)_ |
| `die-roll schedule describe` | `--schedule-id` | _(none)_ |

# Messages
<a id="message_example_v1_ContinueSignalRequest"></a>
## example.v1.ContinueSignalRequest
Instructs the workflow to continue or stop
| Field name | Type | Cardinality | Deprecated ? | Description |
| --- | --- | --- | --- | --- |
| Continue | bool | Optional | ✅ | <pre></pre> |


<a id="message_example_v1_GetStatusResponse"></a>
## example.v1.GetStatusResponse
Returns the progress
| Field name | Type | Cardinality | Deprecated ? | Description |
| --- | --- | --- | --- | --- |
| Progress | int64 | Optional | ✅ | <pre></pre> |


<a id="message_example_v1_ThrowDieResponse"></a>
## example.v1.ThrowDieResponse
Returns the value that was rolled
| Field name | Type | Cardinality | Deprecated ? | Description |
| --- | --- | --- | --- | --- |
| Result | int32 | Optional | ✅ | <pre></pre> |


<a id="message_example_v1_ThrowDiesResponse"></a>
## example.v1.ThrowDiesResponse
Returns the values of a series of rolls
| Field name | Type | Cardinality | Deprecated ? | Description |
| --- | --- | --- | --- | --- |
| Results | int32 | Repeated | ✅ | <pre>Results of the throws</pre> |


<a id="message_example_v1_ThrowDiesRequest"></a>
## example.v1.ThrowDiesRequest
Triggers a series of die rolls
| Field name | Type | Cardinality | Deprecated ? | Description |
| --- | --- | --- | --- | --- |
| Results | int32 | Optional | ✅ | <pre>Result array</pre> |
| Loop | bool | Optional | ✅ | <pre>Loop ?</pre> |
| ResultStatus | string | Optional | 🗿 | <pre>A deprecated field</pre> |


<a id="message_example_v1_ThrowUntilValueRequest"></a>
## example.v1.ThrowUntilValueRequest
Requests  to roll a die until a certain value is pulled
| Field name | Type | Cardinality | Deprecated ? | Description |
| --- | --- | --- | --- | --- |
| Value | int32 | Optional | ✅ | <pre>Target value</pre> |


<a id="message_example_v1_ThrowStatusResponse"></a>
## example.v1.ThrowStatusResponse
Response to a die roll request
| Field name | Type | Cardinality | Deprecated ? | Description |
| --- | --- | --- | --- | --- |
| Throws | int32 | Optional | ✅ | <pre>Number of throws</pre> |


<a id="message_example_v1_ParentWorkflowReply"></a>
## example.v1.ParentWorkflowReply

| Field name | Type | Cardinality | Deprecated ? | Description |
| --- | --- | --- | --- | --- |
| Status | enum | Optional | ✅ | <pre>Status of the workflow</pre> |




[Back to top](#top)
