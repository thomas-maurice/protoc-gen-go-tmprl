// CLI integration tests for the generated Cobra commands. These tests
// execute each generated command via cobra.Command.SetArgs against a fake
// client.Client that captures the calls; the tests assert that the correct
// request message fields, schedule options, and SDK helpers are produced
// from user-supplied flag values. No Temporal server is contacted.
//
// Coverage note: DieRoll's inputs are simple scalars and bools. Exhaustive
// FlagPlan behavior (enums, oneofs, WKT wrappers, optional presence,
// repeated-message JSON escape, maps, nested-message path flattening,
// cycle detection) is verified in `internal/model/flagplan_test.go`
// against synthetic proto descriptors — not duplicated here.
package examplev1

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"google.golang.org/protobuf/proto"
)

// -----------------------------------------------------------------------------
// Fake Temporal client + handle infrastructure.
// -----------------------------------------------------------------------------

// executeCall captures one ExecuteWorkflow invocation against the fake.
type executeCall struct {
	options      client.StartWorkflowOptions
	workflowType string
	args         []interface{}
}

// signalCall captures one SignalWorkflow invocation against the fake.
type signalCall struct {
	workflowID string
	runID      string
	signalName string
	arg        interface{}
}

// queryCall captures one QueryWorkflow invocation against the fake.
type queryCall struct {
	workflowID string
	runID      string
	queryType  string
	args       []interface{}
}

// fakeRun is a minimal client.WorkflowRun capturing no state.
type fakeRun struct {
	id    string
	runID string
	// result is copied into the caller's valuePtr on Get when non-nil so
	// the sync execute path returns a non-nil ThrowDiesResponse.
	result *ThrowDiesResponse
}

func (f *fakeRun) GetID() string    { return f.id }
func (f *fakeRun) GetRunID() string { return f.runID }
func (f *fakeRun) Get(_ context.Context, valuePtr interface{}) error {
	if valuePtr == nil || f.result == nil {
		return nil
	}
	// The Execute helper does `var resp *ThrowDiesResponse; future.Get(ctx, &resp)`;
	// the target is therefore **ThrowDiesResponse. Clone via proto.Clone
	// so we don't copy the embedded MessageState mutex.
	pp, ok := valuePtr.(**ThrowDiesResponse)
	if !ok {
		return errors.New("fakeRun.Get: unexpected valuePtr type")
	}
	*pp = proto.Clone(f.result).(*ThrowDiesResponse)
	return nil
}

func (f *fakeRun) GetWithOptions(ctx context.Context, valuePtr interface{}, _ client.WorkflowRunGetOptions) error {
	return f.Get(ctx, valuePtr)
}

// fakeEncodedValue returns a preset proto message whose bytes are fed into
// the caller's valuePtr on Get.
type fakeEncodedValue struct {
	resp *ThrowStatusResponse
}

func (f *fakeEncodedValue) HasValue() bool { return f.resp != nil }
func (f *fakeEncodedValue) Get(valuePtr interface{}) error {
	if valuePtr == nil {
		return nil
	}
	pp, ok := valuePtr.(**ThrowStatusResponse)
	if !ok {
		return errors.New("fakeEncodedValue.Get: unexpected valuePtr type")
	}
	*pp = proto.Clone(f.resp).(*ThrowStatusResponse)
	return nil
}

// fakeScheduleHandle records the mutating calls the CLI triggers on a handle.
type fakeScheduleHandle struct {
	id             string
	pauseCalls     []client.SchedulePauseOptions
	unpauseCalls   []client.ScheduleUnpauseOptions
	deleteCalls    int
	describeCalls  int
	describeResult *client.ScheduleDescription
	describeErr    error
	pauseErr       error
	unpauseErr     error
	deleteErr      error
}

func (f *fakeScheduleHandle) GetID() string { return f.id }
func (f *fakeScheduleHandle) Delete(ctx context.Context) error {
	f.deleteCalls++
	return f.deleteErr
}
func (f *fakeScheduleHandle) Backfill(ctx context.Context, _ client.ScheduleBackfillOptions) error {
	return nil
}
func (f *fakeScheduleHandle) Update(ctx context.Context, _ client.ScheduleUpdateOptions) error {
	return nil
}
func (f *fakeScheduleHandle) Describe(ctx context.Context) (*client.ScheduleDescription, error) {
	f.describeCalls++
	if f.describeErr != nil {
		return nil, f.describeErr
	}
	if f.describeResult != nil {
		return f.describeResult, nil
	}
	return &client.ScheduleDescription{}, nil
}
func (f *fakeScheduleHandle) Trigger(ctx context.Context, _ client.ScheduleTriggerOptions) error {
	return nil
}
func (f *fakeScheduleHandle) Pause(ctx context.Context, opts client.SchedulePauseOptions) error {
	f.pauseCalls = append(f.pauseCalls, opts)
	return f.pauseErr
}
func (f *fakeScheduleHandle) Unpause(ctx context.Context, opts client.ScheduleUnpauseOptions) error {
	f.unpauseCalls = append(f.unpauseCalls, opts)
	return f.unpauseErr
}

// fakeScheduleClient is the ScheduleClient surface touched by the CLI.
type fakeScheduleClient struct {
	createCalls []client.ScheduleOptions
	handles     map[string]*fakeScheduleHandle
	fallback    *fakeScheduleHandle
}

func newFakeScheduleClient() *fakeScheduleClient {
	return &fakeScheduleClient{
		handles:  map[string]*fakeScheduleHandle{},
		fallback: &fakeScheduleHandle{},
	}
}

func (f *fakeScheduleClient) Create(ctx context.Context, opts client.ScheduleOptions) (client.ScheduleHandle, error) {
	f.createCalls = append(f.createCalls, opts)
	h := f.handleFor(opts.ID)
	return h, nil
}

func (f *fakeScheduleClient) List(ctx context.Context, _ client.ScheduleListOptions) (client.ScheduleListIterator, error) {
	return nil, errors.New("List not supported by fakeScheduleClient")
}

func (f *fakeScheduleClient) GetHandle(ctx context.Context, id string) client.ScheduleHandle {
	return f.handleFor(id)
}

func (f *fakeScheduleClient) handleFor(id string) *fakeScheduleHandle {
	if h, ok := f.handles[id]; ok {
		return h
	}
	f.fallback.id = id
	return f.fallback
}

// fakeClient is the client.Client subset the CLI uses. Methods not set on
// this fake are delegated to an embedded nil interface and will panic if
// ever called — the test asserts which entry points are actually exercised.
type fakeClient struct {
	client.Client // nil embedded interface; methods not explicitly set panic.

	executeCalls    []executeCall
	signalCalls     []signalCall
	queryCalls      []queryCall
	cancelCalls     []struct{ workflowID, runID string }
	terminateCalls  []struct{ workflowID, runID, reason string }
	executeRun      *fakeRun
	executeErr      error
	signalErr       error
	queryResp       *ThrowStatusResponse
	queryErr        error
	scheduleClientV *fakeScheduleClient
}

func newFakeClient() *fakeClient {
	return &fakeClient{
		executeRun:      &fakeRun{id: "wf-123", runID: "run-456"},
		scheduleClientV: newFakeScheduleClient(),
	}
}

func (f *fakeClient) ExecuteWorkflow(ctx context.Context, options client.StartWorkflowOptions, workflowType interface{}, args ...interface{}) (client.WorkflowRun, error) {
	name, _ := workflowType.(string)
	f.executeCalls = append(f.executeCalls, executeCall{
		options:      options,
		workflowType: name,
		args:         args,
	})
	if f.executeErr != nil {
		return nil, f.executeErr
	}
	return f.executeRun, nil
}

func (f *fakeClient) SignalWorkflow(ctx context.Context, workflowID, runID, signalName string, arg interface{}) error {
	f.signalCalls = append(f.signalCalls, signalCall{
		workflowID: workflowID,
		runID:      runID,
		signalName: signalName,
		arg:        arg,
	})
	return f.signalErr
}

func (f *fakeClient) QueryWorkflow(ctx context.Context, workflowID, runID, queryType string, args ...interface{}) (converter.EncodedValue, error) {
	f.queryCalls = append(f.queryCalls, queryCall{
		workflowID: workflowID,
		runID:      runID,
		queryType:  queryType,
		args:       args,
	})
	if f.queryErr != nil {
		return nil, f.queryErr
	}
	resp := f.queryResp
	if resp == nil {
		resp = &ThrowStatusResponse{Throws: 0}
	}
	return &fakeEncodedValue{resp: resp}, nil
}

func (f *fakeClient) CancelWorkflow(ctx context.Context, workflowID, runID string) error {
	f.cancelCalls = append(f.cancelCalls, struct{ workflowID, runID string }{workflowID, runID})
	return nil
}

func (f *fakeClient) TerminateWorkflow(ctx context.Context, workflowID, runID, reason string, _ ...interface{}) error {
	f.terminateCalls = append(f.terminateCalls, struct{ workflowID, runID, reason string }{workflowID, runID, reason})
	return nil
}

func (f *fakeClient) ScheduleClient() client.ScheduleClient {
	return f.scheduleClientV
}

// Note: the generated `workflow describe` command calls
// DescribeWorkflowExecution, which returns
// *workflowservice.DescribeWorkflowExecutionResponse. We intentionally do
// not exercise that command here — its tests would require pulling in the
// workflowservice proto types to build a non-nil response. Schedule
// describe is covered below and exercises the same code path on handles.

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

// runRoot constructs the CLI root command, sets args, captures stdout/stderr,
// and returns stdout, stderr, and the execution error.
func runRoot(t *testing.T, fc *fakeClient, args ...string) (string, string, error) {
	t.Helper()
	root := NewDieRollCLI(fc)
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs(args)
	err := root.Execute()
	return outBuf.String(), errBuf.String(), err
}

// lastExecute asserts the fake captured exactly one ExecuteWorkflow call and
// returns the typed request along with the captured options.
func lastExecute(t *testing.T, fc *fakeClient) (*ThrowDiesRequest, client.StartWorkflowOptions) {
	t.Helper()
	if len(fc.executeCalls) != 1 {
		t.Fatalf("ExecuteWorkflow call count: got %d want 1", len(fc.executeCalls))
	}
	c := fc.executeCalls[0]
	if len(c.args) != 1 {
		t.Fatalf("ExecuteWorkflow args count: got %d want 1", len(c.args))
	}
	req, ok := c.args[0].(*ThrowDiesRequest)
	if !ok {
		t.Fatalf("ExecuteWorkflow arg type: got %T want *ThrowDiesRequest", c.args[0])
	}
	return req, c.options
}

// -----------------------------------------------------------------------------
// Workflow start / execute coverage
// -----------------------------------------------------------------------------

// TestCLIWorkflowStart_WithScalarFields exercises the simple scalar flag
// kinds on `workflow start throw-dies` and checks they land on the
// generated request proto.
func TestCLIWorkflowStart_WithScalarFields(t *testing.T) {
	fc := newFakeClient()
	stdout, _, err := runRoot(t, fc,
		"workflow", "start", "throw-dies",
		"--workflow-id", "my-wf",
		"--results", "7",
		"--loop",
	)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	req, opts := lastExecute(t, fc)
	if opts.ID != "my-wf" {
		t.Errorf("ID = %q, want my-wf", opts.ID)
	}
	if req.GetResults() != 7 {
		t.Errorf("Results = %d, want 7", req.GetResults())
	}
	if !req.GetLoop() {
		t.Error("Loop should be true")
	}
	if !strings.Contains(stdout, "workflow-id: wf-123") {
		t.Errorf("stdout should contain workflow id, got: %q", stdout)
	}
}

// TestCLIWorkflowStart_EmptyInput starts a workflow whose input is
// google.protobuf.Empty; no input flags are registered but targeting and
// schedule-agnostic flags still work.
func TestCLIWorkflowStart_EmptyInput(t *testing.T) {
	fc := newFakeClient()
	_, _, err := runRoot(t, fc,
		"workflow", "start", "parent-workflow",
		"--workflow-id", "parent-1",
	)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if n := len(fc.executeCalls); n != 1 {
		t.Fatalf("ExecuteWorkflow count: got %d want 1", n)
	}
	if fc.executeCalls[0].options.ID != "parent-1" {
		t.Errorf("ID = %q, want parent-1", fc.executeCalls[0].options.ID)
	}
}

// TestCLIWorkflowSignal asserts the signal subcommand sends the signal
// with the correct workflow targeting fields and request proto.
func TestCLIWorkflowSignal(t *testing.T) {
	fc := newFakeClient()
	_, _, err := runRoot(t, fc,
		"workflow", "signal", "throw-dies", "continue",
		"--workflow-id", "wf-sig",
		"--run-id", "run-sig",
		"--continue",
	)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(fc.signalCalls) != 1 {
		t.Fatalf("SignalWorkflow count: got %d want 1", len(fc.signalCalls))
	}
	s := fc.signalCalls[0]
	if s.workflowID != "wf-sig" || s.runID != "run-sig" {
		t.Errorf("target = (%q,%q), want (wf-sig,run-sig)", s.workflowID, s.runID)
	}
	if s.signalName != SignalContinueName {
		t.Errorf("signalName = %q, want %q", s.signalName, SignalContinueName)
	}
	req, ok := s.arg.(*ContinueSignalRequest)
	if !ok {
		t.Fatalf("arg type %T want *ContinueSignalRequest", s.arg)
	}
	if !req.GetContinue() {
		t.Error("signal req.Continue should be true")
	}
}

// TestCLIWorkflowQuery asserts the query subcommand issues QueryWorkflow and
// prints the JSON-encoded response on stdout. GetThrowsStatus takes no
// input parameters; only the targeting flags are registered.
func TestCLIWorkflowQuery(t *testing.T) {
	fc := newFakeClient()
	fc.queryResp = &ThrowStatusResponse{Throws: 42}
	stdout, _, err := runRoot(t, fc,
		"workflow", "query", "throw-until-value", "get-throws-status",
		"--workflow-id", "wf-q",
	)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(fc.queryCalls) != 1 {
		t.Fatalf("QueryWorkflow count: got %d want 1", len(fc.queryCalls))
	}
	q := fc.queryCalls[0]
	if q.workflowID != "wf-q" {
		t.Errorf("query workflowID = %q, want wf-q", q.workflowID)
	}
	if q.queryType != QueryGetThrowsStatusName {
		t.Errorf("queryType = %q, want %q", q.queryType, QueryGetThrowsStatusName)
	}
	if !strings.Contains(stdout, "42") {
		t.Errorf("stdout should contain response Throws count, got: %q", stdout)
	}
}

// TestCLIWorkflowCancelTerminate verifies the plain target-lookup commands.
func TestCLIWorkflowCancelTerminate(t *testing.T) {
	fc := newFakeClient()
	if _, _, err := runRoot(t, fc,
		"workflow", "cancel", "--workflow-id", "wf-c", "--run-id", "run-c",
	); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if len(fc.cancelCalls) != 1 || fc.cancelCalls[0].workflowID != "wf-c" {
		t.Fatalf("cancel calls = %+v", fc.cancelCalls)
	}
	if _, _, err := runRoot(t, fc,
		"workflow", "terminate", "--workflow-id", "wf-t", "--reason", "boom",
	); err != nil {
		t.Fatalf("terminate: %v", err)
	}
	if len(fc.terminateCalls) != 1 || fc.terminateCalls[0].reason != "boom" {
		t.Fatalf("terminate calls = %+v", fc.terminateCalls)
	}
}

// -----------------------------------------------------------------------------
// Schedule create coverage
// -----------------------------------------------------------------------------

// scheduleCreateCommon runs `schedule create throw-dies` with the supplied
// extra args on top of the required --schedule-id and returns the captured
// ScheduleOptions and the created request proto. It also returns the fake
// so callers can make further assertions.
func scheduleCreateCommon(t *testing.T, extraArgs ...string) (*fakeClient, client.ScheduleOptions, *ThrowDiesRequest) {
	t.Helper()
	fc := newFakeClient()
	args := append([]string{
		"schedule", "create", "throw-dies",
		"--schedule-id", "sched-1",
		"--results", "3",
	}, extraArgs...)
	_, _, err := runRoot(t, fc, args...)
	if err != nil {
		t.Fatalf("schedule create: %v", err)
	}
	if n := len(fc.scheduleClientV.createCalls); n != 1 {
		t.Fatalf("schedule Create count: got %d want 1", n)
	}
	opts := fc.scheduleClientV.createCalls[0]
	action, ok := opts.Action.(*client.ScheduleWorkflowAction)
	if !ok {
		t.Fatalf("Action type = %T, want *ScheduleWorkflowAction", opts.Action)
	}
	if len(action.Args) != 1 {
		t.Fatalf("Action.Args len = %d, want 1", len(action.Args))
	}
	req, ok := action.Args[0].(*ThrowDiesRequest)
	if !ok {
		t.Fatalf("Action.Args[0] type = %T, want *ThrowDiesRequest", action.Args[0])
	}
	return fc, opts, req
}

func TestCLIScheduleCreate_Cron(t *testing.T) {
	_, opts, _ := scheduleCreateCommon(t, "--schedule-cron", "0 9 * * *")
	if got := opts.Spec.CronExpressions; len(got) != 1 || got[0] != "0 9 * * *" {
		t.Errorf("CronExpressions = %v, want [0 9 * * *]", got)
	}
}

func TestCLIScheduleCreate_MultipleCron(t *testing.T) {
	_, opts, _ := scheduleCreateCommon(t,
		"--schedule-cron", "0 9 * * *",
		"--schedule-cron", "0 17 * * *",
	)
	want := []string{"0 9 * * *", "0 17 * * *"}
	if len(opts.Spec.CronExpressions) != 2 {
		t.Fatalf("CronExpressions count: got %d want 2", len(opts.Spec.CronExpressions))
	}
	for i, w := range want {
		if opts.Spec.CronExpressions[i] != w {
			t.Errorf("CronExpressions[%d] = %q, want %q", i, opts.Spec.CronExpressions[i], w)
		}
	}
}

func TestCLIScheduleCreate_MultipleIntervals(t *testing.T) {
	_, opts, _ := scheduleCreateCommon(t,
		"--schedule-interval", "5m",
		"--schedule-interval", "1h",
	)
	if n := len(opts.Spec.Intervals); n != 2 {
		t.Fatalf("Intervals count: got %d want 2", n)
	}
	if opts.Spec.Intervals[0].Every != 5*time.Minute {
		t.Errorf("Intervals[0].Every = %v, want 5m", opts.Spec.Intervals[0].Every)
	}
	if opts.Spec.Intervals[1].Every != time.Hour {
		t.Errorf("Intervals[1].Every = %v, want 1h", opts.Spec.Intervals[1].Every)
	}
}

func TestCLIScheduleCreate_StartAtEndAt(t *testing.T) {
	_, opts, _ := scheduleCreateCommon(t,
		"--schedule-start-at", "2025-01-02T15:04:05Z",
		"--schedule-end-at", "2026-01-02T15:04:05Z",
	)
	wantStart := time.Date(2025, 1, 2, 15, 4, 5, 0, time.UTC)
	wantEnd := time.Date(2026, 1, 2, 15, 4, 5, 0, time.UTC)
	if !opts.Spec.StartAt.Equal(wantStart) {
		t.Errorf("StartAt = %v, want %v", opts.Spec.StartAt, wantStart)
	}
	if !opts.Spec.EndAt.Equal(wantEnd) {
		t.Errorf("EndAt = %v, want %v", opts.Spec.EndAt, wantEnd)
	}
}

func TestCLIScheduleCreate_Timezone(t *testing.T) {
	_, opts, _ := scheduleCreateCommon(t, "--schedule-timezone", "Europe/Paris")
	if opts.Spec.TimeZoneName != "Europe/Paris" {
		t.Errorf("TimeZoneName = %q, want Europe/Paris", opts.Spec.TimeZoneName)
	}
}

func TestCLIScheduleCreate_Jitter(t *testing.T) {
	_, opts, _ := scheduleCreateCommon(t, "--schedule-jitter", "30s")
	if opts.Spec.Jitter != 30*time.Second {
		t.Errorf("Jitter = %v, want 30s", opts.Spec.Jitter)
	}
}

func TestCLIScheduleCreate_Paused(t *testing.T) {
	_, opts, _ := scheduleCreateCommon(t, "--schedule-paused")
	if !opts.Paused {
		t.Error("Paused should be true when --schedule-paused is set")
	}
}

func TestCLIScheduleCreate_Note(t *testing.T) {
	_, opts, _ := scheduleCreateCommon(t, "--schedule-note", "bootstrap")
	if opts.Note != "bootstrap" {
		t.Errorf("Note = %q, want bootstrap", opts.Note)
	}
}

// TestCLIScheduleCreate_OverlapKebabNames exercises every accepted overlap
// value against the matching enumspb.ScheduleOverlapPolicy constant.
func TestCLIScheduleCreate_OverlapKebabNames(t *testing.T) {
	cases := []struct {
		cli  string
		want enumspb.ScheduleOverlapPolicy
	}{
		{"skip", enumspb.SCHEDULE_OVERLAP_POLICY_SKIP},
		{"buffer-one", enumspb.SCHEDULE_OVERLAP_POLICY_BUFFER_ONE},
		{"buffer-all", enumspb.SCHEDULE_OVERLAP_POLICY_BUFFER_ALL},
		{"cancel-other", enumspb.SCHEDULE_OVERLAP_POLICY_CANCEL_OTHER},
		{"terminate-other", enumspb.SCHEDULE_OVERLAP_POLICY_TERMINATE_OTHER},
		{"allow-all", enumspb.SCHEDULE_OVERLAP_POLICY_ALLOW_ALL},
	}
	for _, tc := range cases {
		t.Run(tc.cli, func(t *testing.T) {
			_, opts, _ := scheduleCreateCommon(t, "--schedule-overlap", tc.cli)
			if opts.Overlap != tc.want {
				t.Errorf("Overlap = %v, want %v", opts.Overlap, tc.want)
			}
		})
	}
}

// TestCLIScheduleCreate_OverlapInvalid verifies that an unknown overlap
// value is rejected with an error mentioning the accepted values.
func TestCLIScheduleCreate_OverlapInvalid(t *testing.T) {
	fc := newFakeClient()
	_, stderr, err := runRoot(t, fc,
		"schedule", "create", "throw-dies",
		"--schedule-id", "sched-1",
		"--schedule-overlap", "bogus",
	)
	if err == nil {
		t.Fatalf("expected error, stderr=%q", stderr)
	}
	if !strings.Contains(err.Error(), "invalid overlap") &&
		!strings.Contains(stderr, "invalid overlap") {
		t.Errorf("error should mention invalid overlap, got err=%v stderr=%q", err, stderr)
	}
	if !strings.Contains(err.Error(), "skip") && !strings.Contains(stderr, "skip") {
		t.Errorf("error should list accepted values, got err=%v stderr=%q", err, stderr)
	}
	if len(fc.scheduleClientV.createCalls) != 0 {
		t.Errorf("schedule Create should not have been called; got %d calls", len(fc.scheduleClientV.createCalls))
	}
}

func TestCLIScheduleCreate_CatchupWindow(t *testing.T) {
	_, opts, _ := scheduleCreateCommon(t, "--schedule-catchup-window", "1h")
	if opts.CatchupWindow != time.Hour {
		t.Errorf("CatchupWindow = %v, want 1h", opts.CatchupWindow)
	}
}

// TestCLIScheduleCreate_TaskQueueOverride: the --schedule-task-queue flag
// must override the service's default task queue on the emitted action.
func TestCLIScheduleCreate_TaskQueueOverride(t *testing.T) {
	_, opts, _ := scheduleCreateCommon(t, "--schedule-task-queue", "custom-queue")
	action := opts.Action.(*client.ScheduleWorkflowAction)
	if action.TaskQueue != "custom-queue" {
		t.Errorf("Action.TaskQueue = %q, want custom-queue", action.TaskQueue)
	}
}

// TestCLIScheduleCreate_TaskQueueDefault: without --schedule-task-queue the
// service default is used.
func TestCLIScheduleCreate_TaskQueueDefault(t *testing.T) {
	_, opts, _ := scheduleCreateCommon(t)
	action := opts.Action.(*client.ScheduleWorkflowAction)
	if action.TaskQueue != DefaultDieRollTaskQueueName {
		t.Errorf("Action.TaskQueue = %q, want %q",
			action.TaskQueue, DefaultDieRollTaskQueueName)
	}
}

// TestCLIScheduleCreate_WorkflowInputFlag: any non-schedule workflow input
// flag reaches the request proto embedded in the ScheduleWorkflowAction.
// This verifies that the schedule-create command still runs the full
// FlagPlan application pass before handing off to CreateSchedule.
func TestCLIScheduleCreate_WorkflowInputFlag(t *testing.T) {
	_, _, req := scheduleCreateCommon(t,
		"--results", "42",
		"--loop",
	)
	if req.GetResults() != 42 {
		t.Errorf("Results = %d, want 42", req.GetResults())
	}
	if !req.GetLoop() {
		t.Error("Loop should be true")
	}
}

// -----------------------------------------------------------------------------
// Schedule pause / unpause / delete / describe
// -----------------------------------------------------------------------------

func TestCLISchedulePauseWithNote(t *testing.T) {
	fc := newFakeClient()
	fc.scheduleClientV.handles["sched-p"] = &fakeScheduleHandle{id: "sched-p"}
	_, _, err := runRoot(t, fc,
		"schedule", "pause",
		"--schedule-id", "sched-p",
		"--note", "maintenance",
	)
	if err != nil {
		t.Fatalf("schedule pause: %v", err)
	}
	h := fc.scheduleClientV.handles["sched-p"]
	if len(h.pauseCalls) != 1 {
		t.Fatalf("Pause call count: got %d want 1", len(h.pauseCalls))
	}
	if h.pauseCalls[0].Note != "maintenance" {
		t.Errorf("Pause.Note = %q, want maintenance", h.pauseCalls[0].Note)
	}
}

func TestCLIScheduleUnpauseWithNote(t *testing.T) {
	fc := newFakeClient()
	fc.scheduleClientV.handles["sched-u"] = &fakeScheduleHandle{id: "sched-u"}
	_, _, err := runRoot(t, fc,
		"schedule", "unpause",
		"--schedule-id", "sched-u",
		"--note", "resume",
	)
	if err != nil {
		t.Fatalf("schedule unpause: %v", err)
	}
	h := fc.scheduleClientV.handles["sched-u"]
	if len(h.unpauseCalls) != 1 {
		t.Fatalf("Unpause call count: got %d want 1", len(h.unpauseCalls))
	}
	if h.unpauseCalls[0].Note != "resume" {
		t.Errorf("Unpause.Note = %q, want resume", h.unpauseCalls[0].Note)
	}
}

func TestCLIScheduleDelete(t *testing.T) {
	fc := newFakeClient()
	fc.scheduleClientV.handles["sched-d"] = &fakeScheduleHandle{id: "sched-d"}
	_, _, err := runRoot(t, fc,
		"schedule", "delete",
		"--schedule-id", "sched-d",
	)
	if err != nil {
		t.Fatalf("schedule delete: %v", err)
	}
	h := fc.scheduleClientV.handles["sched-d"]
	if h.deleteCalls != 1 {
		t.Errorf("Delete call count: got %d want 1", h.deleteCalls)
	}
}

// TestCLIScheduleDescribe confirms Describe is invoked on the addressed
// handle and its JSON-encoded result is written to stdout.
func TestCLIScheduleDescribe(t *testing.T) {
	fc := newFakeClient()
	fc.scheduleClientV.handles["sched-x"] = &fakeScheduleHandle{
		id: "sched-x",
		describeResult: &client.ScheduleDescription{
			Schedule: client.Schedule{
				State: &client.ScheduleState{
					Note:   "hello from describe",
					Paused: true,
				},
			},
		},
	}
	stdout, _, err := runRoot(t, fc,
		"schedule", "describe",
		"--schedule-id", "sched-x",
	)
	if err != nil {
		t.Fatalf("schedule describe: %v", err)
	}
	h := fc.scheduleClientV.handles["sched-x"]
	if h.describeCalls != 1 {
		t.Errorf("Describe call count: got %d want 1", h.describeCalls)
	}
	if !strings.Contains(stdout, "hello from describe") {
		t.Errorf("stdout should contain the canned Note, got: %q", stdout)
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &decoded); err != nil {
		t.Errorf("stdout should be valid JSON, got err %v, raw: %q", err, stdout)
	}
}

// -----------------------------------------------------------------------------
// Compile-time assertions
// -----------------------------------------------------------------------------

var _ client.ScheduleHandle = (*fakeScheduleHandle)(nil)
var _ client.ScheduleClient = (*fakeScheduleClient)(nil)
var _ proto.Message = (*ThrowDiesRequest)(nil)
