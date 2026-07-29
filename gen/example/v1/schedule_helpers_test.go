// Tests for the generated per-service schedule option helpers. These helpers
// are emitted into the generated file (example_tmprl.pb.go) by the Go-template
// renderer; they live in the same package as the generated code so we can
// exercise them directly without exporting them. If the generator changes the
// shape of these helpers, these tests must be kept in sync.
package examplev1

import (
	"testing"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
)

// newBase returns a ScheduleOptions that mirrors what the generated
// CreateSchedule/UpsertSchedule methods build before merging user input. Each
// test starts from a fresh copy so mutations don't leak across cases.
func newBase() client.ScheduleOptions {
	return client.ScheduleOptions{
		ID: "generator-built-id",
		Spec: client.ScheduleSpec{
			TimeZoneName: "Europe/Paris",
		},
		Action: &client.ScheduleWorkflowAction{
			ID:        "generator-built-id",
			Workflow:  WorkflowProcessOrderName,
			TaskQueue: "generator-tq",
		},
	}
}

// TestMergeScheduleOptionsOrders_EmptyUserLeavesBaseUntouched: verifies that
// passing a zero-valued user ScheduleOptions is a no-op so callers that don't
// supply options get pure generator defaults.
func TestMergeScheduleOptionsOrders_EmptyUserLeavesBaseUntouched(t *testing.T) {
	base := newBase()
	mergeScheduleOptionsOrders(&base, client.ScheduleOptions{}, true)

	if base.Spec.TimeZoneName != "Europe/Paris" {
		t.Errorf("TimeZoneName was clobbered: got %q", base.Spec.TimeZoneName)
	}
	if base.Note != "" {
		t.Errorf("Note was set unexpectedly: got %q", base.Note)
	}
	if base.Paused {
		t.Error("Paused flipped to true from a zero user struct")
	}
	action, ok := base.Action.(*client.ScheduleWorkflowAction)
	if !ok {
		t.Fatalf("Action type changed: %T", base.Action)
	}
	if action.TaskQueue != "generator-tq" {
		t.Errorf("Action.TaskQueue clobbered: got %q", action.TaskQueue)
	}
}

// TestMergeScheduleOptionsOrders_SpecFields: verifies each field on Spec is
// merged when set on the user struct and left alone when zero.
func TestMergeScheduleOptionsOrders_SpecFields(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)

	base := newBase()
	user := client.ScheduleOptions{
		Spec: client.ScheduleSpec{
			CronExpressions: []string{"*/5 * * * *"},
			Calendars:       []client.ScheduleCalendarSpec{{Year: []client.ScheduleRange{{Start: 2026}}}},
			Intervals:       []client.ScheduleIntervalSpec{{Every: time.Minute}},
			Skip:            []client.ScheduleCalendarSpec{{Hour: []client.ScheduleRange{{Start: 3}}}},
			StartAt:         start,
			EndAt:           end,
			Jitter:          30 * time.Second,
			TimeZoneName:    "UTC",
		},
	}

	mergeScheduleOptionsOrders(&base, user, true)

	if got := base.Spec.CronExpressions; len(got) != 1 || got[0] != "*/5 * * * *" {
		t.Errorf("CronExpressions not merged: %v", got)
	}
	if len(base.Spec.Calendars) != 1 {
		t.Errorf("Calendars not merged: %v", base.Spec.Calendars)
	}
	if len(base.Spec.Intervals) != 1 {
		t.Errorf("Intervals not merged: %v", base.Spec.Intervals)
	}
	if len(base.Spec.Skip) != 1 {
		t.Errorf("Skip not merged: %v", base.Spec.Skip)
	}
	if !base.Spec.StartAt.Equal(start) {
		t.Errorf("StartAt = %v, want %v", base.Spec.StartAt, start)
	}
	if !base.Spec.EndAt.Equal(end) {
		t.Errorf("EndAt = %v, want %v", base.Spec.EndAt, end)
	}
	if base.Spec.Jitter != 30*time.Second {
		t.Errorf("Jitter = %v, want 30s", base.Spec.Jitter)
	}
	if base.Spec.TimeZoneName != "UTC" {
		t.Errorf("TimeZoneName = %q, want UTC", base.Spec.TimeZoneName)
	}
}

// TestMergeScheduleOptionsOrders_TopLevelFields: verifies non-Spec fields
// (Note, Overlap, CatchupWindow, PauseOnFailure, Action) merge correctly.
func TestMergeScheduleOptionsOrders_TopLevelFields(t *testing.T) {
	userAction := &client.ScheduleWorkflowAction{
		ID:        "user-action-id",
		Workflow:  WorkflowProcessOrderName,
		TaskQueue: "user-tq",
	}

	base := newBase()
	user := client.ScheduleOptions{
		Note:           "hello",
		Overlap:        3, // arbitrary non-zero enum value
		CatchupWindow:  2 * time.Minute,
		PauseOnFailure: true,
		Action:         userAction,
	}

	mergeScheduleOptionsOrders(&base, user, true)

	if base.Note != "hello" {
		t.Errorf("Note = %q, want hello", base.Note)
	}
	if base.Overlap != 3 {
		t.Errorf("Overlap = %v, want 3", base.Overlap)
	}
	if base.CatchupWindow != 2*time.Minute {
		t.Errorf("CatchupWindow = %v, want 2m", base.CatchupWindow)
	}
	if !base.PauseOnFailure {
		t.Error("PauseOnFailure should be true")
	}
	if base.Action != userAction {
		t.Error("Action pointer not replaced by user-supplied one")
	}
}

// TestMergeScheduleOptionsOrders_PausedGatedByApplyPaused: the applyPaused
// flag is the only behavioural difference between the Create and Upsert call
// sites. Create passes true (user can create a schedule in the paused state);
// Upsert passes false so that a partial options struct never silently pauses
// an already-running schedule.
func TestMergeScheduleOptionsOrders_PausedGatedByApplyPaused(t *testing.T) {
	t.Run("applyPaused=true honours user Paused", func(t *testing.T) {
		base := newBase()
		mergeScheduleOptionsOrders(&base, client.ScheduleOptions{Paused: true}, true)
		if !base.Paused {
			t.Error("Paused should be true when applyPaused is true")
		}
	})

	t.Run("applyPaused=false ignores user Paused", func(t *testing.T) {
		base := newBase()
		mergeScheduleOptionsOrders(&base, client.ScheduleOptions{Paused: true}, false)
		if base.Paused {
			t.Error("Paused should stay false when applyPaused is false")
		}
	})

	t.Run("applyPaused=true does not force Paused when user left it false", func(t *testing.T) {
		base := newBase()
		base.Paused = false
		mergeScheduleOptionsOrders(&base, client.ScheduleOptions{Paused: false}, true)
		if base.Paused {
			t.Error("Paused should remain false when user did not set it")
		}
	})
}

// TestMergeScheduleOptionsOrders_TypedSearchAttributes: verifies the typed
// search attributes collection is replaced wholesale when the user provides a
// non-empty one and left alone when empty.
func TestMergeScheduleOptionsOrders_TypedSearchAttributes(t *testing.T) {
	key := temporal.NewSearchAttributeKeyString("CustomKey")

	base := newBase()
	user := client.ScheduleOptions{
		TypedSearchAttributes: temporal.NewSearchAttributes(key.ValueSet("user-value")),
	}
	mergeScheduleOptionsOrders(&base, user, false)

	got, ok := base.TypedSearchAttributes.GetString(key)
	if !ok || got != "user-value" {
		t.Errorf("TypedSearchAttributes not merged, got=%q ok=%v", got, ok)
	}

	// Empty user attrs must NOT wipe the base's attrs.
	base2 := newBase()
	base2.TypedSearchAttributes = temporal.NewSearchAttributes(key.ValueSet("base-value"))
	mergeScheduleOptionsOrders(&base2, client.ScheduleOptions{}, false)
	got2, ok2 := base2.TypedSearchAttributes.GetString(key)
	if !ok2 || got2 != "base-value" {
		t.Errorf("empty user attrs clobbered base attrs, got=%q ok=%v", got2, ok2)
	}
}

// TestApplyScheduleDefaultsOrders_FillsTaskQueueWhenEmpty: the service's
// default task queue is applied only when the ScheduleWorkflowAction has an
// empty TaskQueue; a user-supplied one always wins.
func TestApplyScheduleDefaultsOrders_FillsTaskQueueWhenEmpty(t *testing.T) {
	t.Run("empty task queue gets default", func(t *testing.T) {
		opts := client.ScheduleOptions{
			Action: &client.ScheduleWorkflowAction{},
		}
		applyScheduleDefaultsOrders(&opts, "default-tq")
		action := opts.Action.(*client.ScheduleWorkflowAction)
		if action.TaskQueue != "default-tq" {
			t.Errorf("TaskQueue = %q, want default-tq", action.TaskQueue)
		}
	})

	t.Run("non-empty task queue is preserved", func(t *testing.T) {
		opts := client.ScheduleOptions{
			Action: &client.ScheduleWorkflowAction{TaskQueue: "user-tq"},
		}
		applyScheduleDefaultsOrders(&opts, "default-tq")
		action := opts.Action.(*client.ScheduleWorkflowAction)
		if action.TaskQueue != "user-tq" {
			t.Errorf("TaskQueue = %q, want user-tq (user override must win)", action.TaskQueue)
		}
	})

	t.Run("non-ScheduleWorkflowAction is left alone", func(t *testing.T) {
		// Pass a nil Action; the type assertion fails and the helper must
		// not panic.
		opts := client.ScheduleOptions{}
		applyScheduleDefaultsOrders(&opts, "default-tq")
		if opts.Action != nil {
			t.Errorf("Action = %v, want nil (no synthesis)", opts.Action)
		}
	})
}

// TestScheduleHelpers_CreateAndUpsertSemantics: documents the behavioural
// difference between Create (applyPaused=true) and Upsert (applyPaused=false)
// by running them back-to-back on the same base and user input.
func TestScheduleHelpers_CreateAndUpsertSemantics(t *testing.T) {
	user := client.ScheduleOptions{
		Note:   "common",
		Paused: true,
	}

	createBase := newBase()
	mergeScheduleOptionsOrders(&createBase, user, true)
	applyScheduleDefaultsOrders(&createBase, "default-tq")

	upsertBase := newBase()
	mergeScheduleOptionsOrders(&upsertBase, user, false)
	applyScheduleDefaultsOrders(&upsertBase, "default-tq")

	if !createBase.Paused {
		t.Error("Create flow should honour Paused=true from user")
	}
	if upsertBase.Paused {
		t.Error("Upsert flow should NOT apply Paused from user input")
	}
	if createBase.Note != upsertBase.Note {
		t.Errorf("Note differs between flows: create=%q upsert=%q", createBase.Note, upsertBase.Note)
	}
}

// TestScheduleSpecIsZeroOrders verifies the helper UpsertSchedule uses to decide
// whether the caller supplied a spec. This is the guard that prevents an upsert
// with no spec from blanking (and thereby stopping) an existing schedule
// (regression for the F1 CRITICAL finding). It must recognise every spec field
// that mergeScheduleOptionsOrders copies from user options.
func TestScheduleSpecIsZeroOrders(t *testing.T) {
	if !scheduleSpecIsZeroOrders(client.ScheduleSpec{}) {
		t.Error("an empty ScheduleSpec must be reported as zero")
	}

	cases := []struct {
		name string
		spec client.ScheduleSpec
	}{
		{"cron", client.ScheduleSpec{CronExpressions: []string{"@daily"}}},
		{"calendars", client.ScheduleSpec{Calendars: []client.ScheduleCalendarSpec{{}}}},
		{"intervals", client.ScheduleSpec{Intervals: []client.ScheduleIntervalSpec{{Every: time.Hour}}}},
		{"skip", client.ScheduleSpec{Skip: []client.ScheduleCalendarSpec{{}}}},
		{"jitter", client.ScheduleSpec{Jitter: time.Minute}},
		{"timezone", client.ScheduleSpec{TimeZoneName: "Europe/Paris"}},
		{"startAt", client.ScheduleSpec{StartAt: time.Now()}},
		{"endAt", client.ScheduleSpec{EndAt: time.Now()}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if scheduleSpecIsZeroOrders(tc.spec) {
				t.Errorf("a ScheduleSpec with %s set must NOT be reported as zero", tc.name)
			}
		})
	}
}
