package testutil

import (
	"reflect"
	"testing"
	"testing/quick"
)

// PropertyConfig defines the configuration for a property test.
type PropertyConfig struct {
	MaxCount int
	Values   func([]reflect.Value, *testing.T)
}

// RunPropertyTest runs a property test with the given function and configuration.
func RunPropertyTest(t *testing.T, feature string, propertyID int, description string, f interface{}, config *quick.Config) {
	t.Helper()
	t.Logf("Running Property Test - Feature: %s, Property %d: %s", feature, propertyID, description)

	if config == nil {
		config = &quick.Config{MaxCount: 100}
	} else if config.MaxCount == 0 {
		config.MaxCount = 100
	}

	if err := quick.Check(f, config); err != nil {
		t.Errorf("Property Test Failed - Feature: %s, Property %d: %v", feature, propertyID, err)
	}
}

// CheckProperty is a helper that wraps quick.Check with default chexum project standards.
func CheckProperty(t *testing.T, f interface{}) {
	t.Helper()
	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(f, config); err != nil {
		t.Error(err)
	}
}
