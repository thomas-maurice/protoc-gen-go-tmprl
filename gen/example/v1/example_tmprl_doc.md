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
| Workflow execution timeout | 24h0m0s |
| Workflow run timeout | 2h0m0s |

### Workflows
<a id="method_example_v1_DieRoll_ParentWorkflow"></a>
#### example.v1.DieRoll.ParentWorkflow
Parent workflow that calls the Child workflow -- to test workflow ID generations mainly

Input: [google.protobuf.Empty](#message_google_protobuf_Empty)

Output: [example.v1.ParentWorkflowReply](#message_example_v1_ParentWorkflowReply)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `example.v1.DieRoll.ParentWorkflow` |


Signals:
 * [example.v1.DieRoll.Continue](#method_example_v1_DieRoll_Continue)

<a id="method_example_v1_DieRoll_ChildWorkflow"></a>
#### example.v1.DieRoll.ChildWorkflow


Input: [google.protobuf.Empty](#message_google_protobuf_Empty)

Output: [google.protobuf.Empty](#message_google_protobuf_Empty)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `example.v1.DieRoll.ChildWorkflow` |


<a id="method_example_v1_DieRoll_ThrowDies"></a>
#### example.v1.DieRoll.ThrowDies
Throws dies a few times and return the result

Input: [example.v1.ThrowDiesRequest](#message_example_v1_ThrowDiesRequest)

Output: [example.v1.ThrowDiesResponse](#message_example_v1_ThrowDiesResponse)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `example.v1.DieRoll.ThrowDies` |


Signals:
 * [example.v1.DieRoll.Continue](#method_example_v1_DieRoll_Continue)

<a id="method_example_v1_DieRoll_ThrowUntilValue"></a>
#### example.v1.DieRoll.ThrowUntilValue


Input: [example.v1.ThrowUntilValueRequest](#message_example_v1_ThrowUntilValueRequest)

Output: [google.protobuf.Empty](#message_google_protobuf_Empty)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `example.v1.DieRoll.ThrowUntilValue` |


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
| Non retryable error types | [FATAL NOT_FOUND] |


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
| Heartbeat timeout | 1m0s |

### Queries
<a id="method_example_v1_DieRoll_GetThrowsStatus"></a>
#### example.v1.DieRoll.GetThrowsStatus
Query the state of a workflow
Query the state of the workflow

Input: [google.protobuf.Empty](#message_google_protobuf_Empty)

Output: [example.v1.ThrowStatusResponse](#message_example_v1_ThrowStatusResponse)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `example.v1.DieRoll.GetThrowsStatus` |

### Signals
<a id="method_example_v1_DieRoll_Continue"></a>
#### example.v1.DieRoll.Continue
Signals can be defined with whatever return type you want as they
 do not expect an answer
Instruct the workflow to proceed

Input: [example.v1.ContinueSignalRequest](#message_example_v1_ContinueSignalRequest)

Output: [google.protobuf.Empty](#message_google_protobuf_Empty)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `example.v1.DieRoll.Continue` |

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
