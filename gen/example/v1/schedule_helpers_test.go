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
			Workflow:  WorkflowThrowDiesName,
			TaskQueue: "generator-tq",
		},
	}
}

// TestMergeScheduleOptionsDieRoll_EmptyUserLeavesBaseUntouched: verifies that
// passing a zero-valued user ScheduleOptions is a no-op so callers that don't
// supply options get pure generator defaults.
func TestMergeScheduleOptionsDieRoll_EmptyUserLeavesBaseUntouched(t *testing.T) {
	base := newBase()
	mergeScheduleOptionsDieRoll(&base, client.ScheduleOptions{}, true)

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

// TestMergeScheduleOptionsDieRoll_SpecFields: verifies each field on Spec is
// merged when set on the user struct and left alone when zero.
func TestMergeScheduleOptionsDieRoll_SpecFields(t *testing.T) {
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

	mergeScheduleOptionsDieRoll(&base, user, true)

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

// TestMergeScheduleOptionsDieRoll_TopLevelFields: verifies non-Spec fields
// (Note, Overlap, CatchupWindow, PauseOnFailure, Action) merge correctly.
func TestMergeScheduleOptionsDieRoll_TopLevelFields(t *testing.T) {
	userAction := &client.ScheduleWorkflowAction{
		ID:        "user-action-id",
		Workflow:  WorkflowThrowDiesName,
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

	mergeScheduleOptionsDieRoll(&base, user, true)

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

// TestMergeScheduleOptionsDieRoll_PausedGatedByApplyPaused: the applyPaused
// flag is the only behavioural difference between the Create and Upsert call
// sites. Create passes true (user can create a schedule in the paused state);
// Upsert passes false so that a partial options struct never silently pauses
// an already-running schedule.
func TestMergeScheduleOptionsDieRoll_PausedGatedByApplyPaused(t *testing.T) {
	t.Run("applyPaused=true honours user Paused", func(t *testing.T) {
		base := newBase()
		mergeScheduleOptionsDieRoll(&base, client.ScheduleOptions{Paused: true}, true)
		if !base.Paused {
			t.Error("Paused should be true when applyPaused is true")
		}
	})

	t.Run("applyPaused=false ignores user Paused", func(t *testing.T) {
		base := newBase()
		mergeScheduleOptionsDieRoll(&base, client.ScheduleOptions{Paused: true}, false)
		if base.Paused {
			t.Error("Paused should stay false when applyPaused is false")
		}
	})

	t.Run("applyPaused=true does not force Paused when user left it false", func(t *testing.T) {
		base := newBase()
		base.Paused = false
		mergeScheduleOptionsDieRoll(&base, client.ScheduleOptions{Paused: false}, true)
		if base.Paused {
			t.Error("Paused should remain false when user did not set it")
		}
	})
}

// TestMergeScheduleOptionsDieRoll_TypedSearchAttributes: verifies the typed
// search attributes collection is replaced wholesale when the user provides a
// non-empty one and left alone when empty.
func TestMergeScheduleOptionsDieRoll_TypedSearchAttributes(t *testing.T) {
	key := temporal.NewSearchAttributeKeyString("CustomKey")

	base := newBase()
	user := client.ScheduleOptions{
		TypedSearchAttributes: temporal.NewSearchAttributes(key.ValueSet("user-value")),
	}
	mergeScheduleOptionsDieRoll(&base, user, false)

	got, ok := base.TypedSearchAttributes.GetString(key)
	if !ok || got != "user-value" {
		t.Errorf("TypedSearchAttributes not merged, got=%q ok=%v", got, ok)
	}

	// Empty user attrs must NOT wipe the base's attrs.
	base2 := newBase()
	base2.TypedSearchAttributes = temporal.NewSearchAttributes(key.ValueSet("base-value"))
	mergeScheduleOptionsDieRoll(&base2, client.ScheduleOptions{}, false)
	got2, ok2 := base2.TypedSearchAttributes.GetString(key)
	if !ok2 || got2 != "base-value" {
		t.Errorf("empty user attrs clobbered base attrs, got=%q ok=%v", got2, ok2)
	}
}

// TestApplyScheduleDefaultsDieRoll_FillsCronWhenSpecEmpty: the generator's
// default cron expression should be copied onto the options whenever the user
// hasn't supplied any Spec entry.
func TestApplyScheduleDefaultsDieRoll_FillsCronWhenSpecEmpty(t *testing.T) {
	opts := client.ScheduleOptions{
		Action: &client.ScheduleWorkflowAction{TaskQueue: "x"},
	}
	applyScheduleDefaultsDieRoll(&opts, "0 * * * *", "default-tq")

	if got := opts.Spec.CronExpressions; len(got) != 1 || got[0] != "0 * * * *" {
		t.Errorf("CronExpressions = %v, want [0 * * * *]", got)
	}
}

// TestApplyScheduleDefaultsDieRoll_PreservesUserSpec: if the user already
// populated any of CronExpressions/Calendars/Intervals the default cron must
// not be appended.
func TestApplyScheduleDefaultsDieRoll_PreservesUserSpec(t *testing.T) {
	cases := map[string]client.ScheduleSpec{
		"CronExpressions set": {CronExpressions: []string{"*/1 * * * *"}},
		"Calendars set":       {Calendars: []client.ScheduleCalendarSpec{{Hour: []client.ScheduleRange{{Start: 9}}}}},
		"Intervals set":       {Intervals: []client.ScheduleIntervalSpec{{Every: time.Hour}}},
	}
	for name, spec := range cases {
		t.Run(name, func(t *testing.T) {
			opts := client.ScheduleOptions{
				Spec:   spec,
				Action: &client.ScheduleWorkflowAction{TaskQueue: "x"},
			}
			applyScheduleDefaultsDieRoll(&opts, "should-not-appear", "default-tq")

			if len(opts.Spec.CronExpressions) == 1 && opts.Spec.CronExpressions[0] == "should-not-appear" {
				t.Error("default cron overwrote user-provided spec")
			}
		})
	}
}

// TestApplyScheduleDefaultsDieRoll_FillsTaskQueueWhenEmpty: the service's
// default task queue is applied only when the ScheduleWorkflowAction has an
// empty TaskQueue; a user-supplied one always wins.
func TestApplyScheduleDefaultsDieRoll_FillsTaskQueueWhenEmpty(t *testing.T) {
	t.Run("empty task queue gets default", func(t *testing.T) {
		opts := client.ScheduleOptions{
			Action: &client.ScheduleWorkflowAction{},
		}
		applyScheduleDefaultsDieRoll(&opts, "* * * * *", "default-tq")
		action := opts.Action.(*client.ScheduleWorkflowAction)
		if action.TaskQueue != "default-tq" {
			t.Errorf("TaskQueue = %q, want default-tq", action.TaskQueue)
		}
	})

	t.Run("non-empty task queue is preserved", func(t *testing.T) {
		opts := client.ScheduleOptions{
			Action: &client.ScheduleWorkflowAction{TaskQueue: "user-tq"},
		}
		applyScheduleDefaultsDieRoll(&opts, "* * * * *", "default-tq")
		action := opts.Action.(*client.ScheduleWorkflowAction)
		if action.TaskQueue != "user-tq" {
			t.Errorf("TaskQueue = %q, want user-tq (user override must win)", action.TaskQueue)
		}
	})

	t.Run("non-ScheduleWorkflowAction is left alone", func(t *testing.T) {
		// Pass a nil Action; the type assertion fails and the helper must
		// not panic.
		opts := client.ScheduleOptions{}
		applyScheduleDefaultsDieRoll(&opts, "* * * * *", "default-tq")
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
	mergeScheduleOptionsDieRoll(&createBase, user, true)
	applyScheduleDefaultsDieRoll(&createBase, "* * * * *", "default-tq")

	upsertBase := newBase()
	mergeScheduleOptionsDieRoll(&upsertBase, user, false)
	applyScheduleDefaultsDieRoll(&upsertBase, "* * * * *", "default-tq")

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
