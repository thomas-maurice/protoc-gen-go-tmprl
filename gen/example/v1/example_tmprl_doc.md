<a id="top"></a>
# Services
<a id="service:example.v1.DieRoll"></a>
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
| Option | Value |
| --- | --- |
| Default task queue | `service-task-queue` |

### Table of contents

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
### Workflows
<a id="method_example_v1_DieRoll_ParentWorkflow"></a>
#### example.v1.DieRoll.ParentWorkflow
Parent workflow that calls the Child workflow -- to test workflow ID generations mainly

Input : [google.protobuf.Empty](#message_google_protobuf_Empty)

Output : [example.v1.ParentWorkflowReply](#message_example_v1_ParentWorkflowReply)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `example.v1.DieRoll.ParentWorkflow` |


Signals:
 * [example.v1.DieRoll.Continue](#method_example_v1_DieRoll_Continue)

<a id="method_example_v1_DieRoll_ChildWorkflow"></a>
#### example.v1.DieRoll.ChildWorkflow


Input : [google.protobuf.Empty](#message_google_protobuf_Empty)

Output : [google.protobuf.Empty](#message_google_protobuf_Empty)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `example.v1.DieRoll.ChildWorkflow` |


<a id="method_example_v1_DieRoll_ThrowDies"></a>
#### example.v1.DieRoll.ThrowDies
Throws dies a few times and return the result

Input : [example.v1.ThrowDiesRequest](#message_example_v1_ThrowDiesRequest)

Output : [example.v1.ThrowDiesResponse](#message_example_v1_ThrowDiesResponse)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `example.v1.DieRoll.ThrowDies` |


Signals:
 * [example.v1.DieRoll.Continue](#method_example_v1_DieRoll_Continue)

<a id="method_example_v1_DieRoll_ThrowUntilValue"></a>
#### example.v1.DieRoll.ThrowUntilValue


Input : [example.v1.ThrowUntilValueRequest](#message_example_v1_ThrowUntilValueRequest)

Output : [google.protobuf.Empty](#message_google_protobuf_Empty)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `example.v1.DieRoll.ThrowUntilValue` |


Queries:

### Activities
<a id="method_example_v1_DieRoll_ThrowDie"></a>
#### example.v1.DieRoll.ThrowDie
Throws a d6 and returns the result

Input : [google.protobuf.Empty](#message_google_protobuf_Empty)

Output : [example.v1.ThrowDieResponse](#message_example_v1_ThrowDieResponse)


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
| Maximum attemps | 10 |
| Maximum interval | 10s |
| Non retryable error types | [FATAL] |


<a id="method_example_v1_DieRoll_Ping"></a>
#### example.v1.DieRoll.Ping
Just a simple ping
 Takes no parameters
 returns nothing

Input : [google.protobuf.Empty](#message_google_protobuf_Empty)

Output : [google.protobuf.Empty](#message_google_protobuf_Empty)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `ping.Ping` |

### Queries
<a id="method_example_v1_DieRoll_GetThrowsStatus"></a>
#### example.v1.DieRoll.GetThrowsStatus
Query the state of a workflow
Query the state of the workflow

Input : [google.protobuf.Empty](#message_google_protobuf_Empty)

Output : [example.v1.ThrowStatusResponse](#message_example_v1_ThrowStatusResponse)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `example.v1.DieRoll.GetThrowsStatus` |

### Signals
<a id="method_example_v1_DieRoll_Continue"></a>
#### example.v1.DieRoll.Continue
Signals can be defined with whatever return type you want as they
 do not expect an answer
Instruct the workflow to proceed

Input : [example.v1.ContinueSignalRequest](#message_example_v1_ContinueSignalRequest)

Output : [example.v1.ContinueSignalRequest](#message_example_v1_ContinueSignalRequest)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `example.v1.DieRoll.Continue` |

# Messages
<a id="message_example_v1_ContinueSignalRequest"></a>
## example.v1.ContinueSignalRequest
Instructs the workflow to continue or stop
<a id="message_example_v1_GetStatusResponse"></a>
## example.v1.GetStatusResponse
Returns the progress
<a id="message_example_v1_ThrowDieResponse"></a>
## example.v1.ThrowDieResponse
Returns the value that was rolled
<a id="message_example_v1_ThrowDiesResponse"></a>
## example.v1.ThrowDiesResponse
Returns the values of a series of rolls
<a id="message_example_v1_ThrowDiesRequest"></a>
## example.v1.ThrowDiesRequest
Triggers a series of die rolls
<a id="message_example_v1_ThrowUntilValueRequest"></a>
## example.v1.ThrowUntilValueRequest
Requests  to roll a die until a certain value is pulled
<a id="message_example_v1_ThrowStatusResponse"></a>
## example.v1.ThrowStatusResponse
Response to a die roll request
<a id="message_example_v1_ParentWorkflowReply"></a>
## example.v1.ParentWorkflowReply

[Back to top](#top)
