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
   * [example.v1.DieRoll.ParentWorkflow](#method:example.v1.DieRoll.ParentWorkflow)
   * [example.v1.DieRoll.ChildWorkflow](#method:example.v1.DieRoll.ChildWorkflow)
   * [example.v1.DieRoll.ThrowDies](#method:example.v1.DieRoll.ThrowDies)
   * [example.v1.DieRoll.ThrowUntilValue](#method:example.v1.DieRoll.ThrowUntilValue)
 * Activities
   * [example.v1.DieRoll.ThrowDie](#method:example.v1.DieRoll.ThrowDie)
   * [example.v1.DieRoll.Ping](#method:example.v1.DieRoll.Ping)
 * Signals
   * [example.v1.DieRoll.Continue](#method:example.v1.DieRoll.Continue)
 * Queries
   * [example.v1.DieRoll.GetThrowsStatus](#method:example.v1.DieRoll.GetThrowsStatus)
### Workflows
<a id="method:example.v1.DieRoll.ParentWorkflow"></a>
#### example.v1.DieRoll.ParentWorkflow
Parent workflow that calls the Child workflow -- to test workflow ID generations mainly

Input : [google.protobuf.Empty](#message:google.protobuf.Empty)

Output : [example.v1.ParentWorkflowReply](#message:example.v1.ParentWorkflowReply)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `example.v1.DieRoll.ParentWorkflow` |


Signals:
 * [example.v1.DieRoll.Continue](#method:example.v1.DieRoll.Continue)

<a id="method:example.v1.DieRoll.ChildWorkflow"></a>
#### example.v1.DieRoll.ChildWorkflow


Input : [google.protobuf.Empty](#message:google.protobuf.Empty)

Output : [google.protobuf.Empty](#message:google.protobuf.Empty)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `example.v1.DieRoll.ChildWorkflow` |


<a id="method:example.v1.DieRoll.ThrowDies"></a>
#### example.v1.DieRoll.ThrowDies
Throws dies a few times and return the result

Input : [example.v1.ThrowDiesRequest](#message:example.v1.ThrowDiesRequest)

Output : [example.v1.ThrowDiesResponse](#message:example.v1.ThrowDiesResponse)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `example.v1.DieRoll.ThrowDies` |


Signals:
 * [example.v1.DieRoll.Continue](#method:example.v1.DieRoll.Continue)

<a id="method:example.v1.DieRoll.ThrowUntilValue"></a>
#### example.v1.DieRoll.ThrowUntilValue


Input : [example.v1.ThrowUntilValueRequest](#message:example.v1.ThrowUntilValueRequest)

Output : [google.protobuf.Empty](#message:google.protobuf.Empty)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `example.v1.DieRoll.ThrowUntilValue` |


Queries:

### Activities
<a id="method:example.v1.DieRoll.ThrowDie"></a>
#### example.v1.DieRoll.ThrowDie
Throws a d6 and returns the result

Input : [google.protobuf.Empty](#message:google.protobuf.Empty)

Output : [example.v1.ThrowDieResponse](#message:example.v1.ThrowDieResponse)


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


<a id="method:example.v1.DieRoll.Ping"></a>
#### example.v1.DieRoll.Ping
Just a simple ping
 Takes no parameters
 returns nothing

Input : [google.protobuf.Empty](#message:google.protobuf.Empty)

Output : [google.protobuf.Empty](#message:google.protobuf.Empty)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `ping.Ping` |

### Queries
<a id="method:example.v1.DieRoll.GetThrowsStatus"></a>
#### example.v1.DieRoll.GetThrowsStatus
Query the state of a workflow
Query the state of the workflow

Input : [google.protobuf.Empty](#message:google.protobuf.Empty)

Output : [example.v1.ThrowStatusResponse](#message:example.v1.ThrowStatusResponse)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `example.v1.DieRoll.GetThrowsStatus` |

### Signals
<a id="method:example.v1.DieRoll.Continue"></a>
#### example.v1.DieRoll.Continue
Signals can be defined with whatever return type you want as they
 do not expect an answer
Instruct the workflow to proceed

Input : [example.v1.ContinueSignalRequest](#message:example.v1.ContinueSignalRequest)

Output : [example.v1.ContinueSignalRequest](#message:example.v1.ContinueSignalRequest)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `example.v1.DieRoll.Continue` |

# Messages
<a id="message:example.v1.ContinueSignalRequest"></a>
## example.v1.ContinueSignalRequest
Instructs the workflow to continue or stop
<a id="message:example.v1.GetStatusResponse"></a>
## example.v1.GetStatusResponse
Returns the progress
<a id="message:example.v1.ThrowDieResponse"></a>
## example.v1.ThrowDieResponse
Returns the value that was rolled
<a id="message:example.v1.ThrowDiesResponse"></a>
## example.v1.ThrowDiesResponse
Returns the values of a series of rolls
<a id="message:example.v1.ThrowDiesRequest"></a>
## example.v1.ThrowDiesRequest
Triggers a series of die rolls
<a id="message:example.v1.ThrowUntilValueRequest"></a>
## example.v1.ThrowUntilValueRequest
Requests  to roll a die until a certain value is pulled
<a id="message:example.v1.ThrowStatusResponse"></a>
## example.v1.ThrowStatusResponse
Response to a die roll request
<a id="message:example.v1.ParentWorkflowReply"></a>
## example.v1.ParentWorkflowReply

[Back to top](#top)
