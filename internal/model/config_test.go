package model

import (
	"testing"
)

// TestConfig: Tests Config struct creation and field access
func TestConfig(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		config := &Config{}

		if config.GenWorkflowPrefix != false {
			t.Error("expected GenWorkflowPrefix to default to false")
		}

		if config.DefaultActivityScheduleToClose != 0 {
			t.Error("expected DefaultActivityScheduleToClose to default to 0")
		}
	})

	t.Run("configured values", func(t *testing.T) {
		config := &Config{
			GenWorkflowPrefix:              true,
			DefaultActivityScheduleToClose: 3600,
		}

		if !config.GenWorkflowPrefix {
			t.Error("expected GenWorkflowPrefix to be true")
		}

		if config.DefaultActivityScheduleToClose != 3600 {
			t.Errorf("expected DefaultActivityScheduleToClose to be 3600, got %d", config.DefaultActivityScheduleToClose)
		}
	})

	t.Run("negative timeout", func(t *testing.T) {
		config := &Config{
			DefaultActivityScheduleToClose: -100,
		}

		if config.DefaultActivityScheduleToClose != -100 {
			t.Errorf("expected DefaultActivityScheduleToClose to be -100, got %d", config.DefaultActivityScheduleToClose)
		}
	})
}
