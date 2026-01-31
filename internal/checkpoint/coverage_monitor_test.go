package checkpoint

import (
	"strings"
	"testing"

	"github.com/Les-El/chexum/internal/testutil"
)

// Property 2: Coverage Threshold Maintenance
// **Validates: Requirements 1.4**
func TestProperty_CoverageThresholdMaintenance(t *testing.T) {
	f := func(coverage float64, threshold float64) bool {
		if threshold <= 0 {
			threshold = 85.0
		}
		monitor := NewCoverageMonitor(threshold)

		data := map[string]float64{
			"github.com/user/pkg": coverage,
		}

		failures, ok := monitor.ValidateThreshold(data)

		if coverage < threshold {
			return !ok && len(failures) == 1
		}
		return ok && len(failures) == 0
	}

	testutil.CheckProperty(t, f)
}

// Property 3: CI Coverage Monitoring
// **Validates: Requirements 1.5**
func TestProperty_CICoverageMonitoring(t *testing.T) {
	f := func(output string) bool {
		monitor := NewCoverageMonitor(85.0)
		coverage, err := monitor.ParseCoverageOutput(output)
		return err == nil && coverage != nil
	}

	testutil.CheckProperty(t, f)
}

func TestCoverageMonitor_ParseOutput(t *testing.T) {
	output := `
ok  	github.com/Les-El/chexum/internal/config	0.015s	coverage: 87.5% of statements
ok  	github.com/Les-El/chexum/internal/hash	0.020s	coverage: 92.0% of statements
ok  	github.com/Les-El/chexum/internal/errors	0.005s	coverage: 45.0% of statements
`
	monitor := NewCoverageMonitor(85.0)
	coverage, err := monitor.ParseCoverageOutput(output)
	if err != nil {
		t.Fatalf("ParseCoverageOutput failed: %v", err)
	}

	if coverage["github.com/Les-El/chexum/internal/config"] != 87.5 {
		t.Errorf("Expected 87.5%%, got %.1f%%", coverage["github.com/Les-El/chexum/internal/config"])
	}

	failures, ok := monitor.ValidateThreshold(coverage)
	if ok {
		t.Error("Expected validation to fail due to internal/errors")
	}
	if len(failures) != 1 {
		t.Errorf("Expected 1 failure, got %d", len(failures))
	}
}

func TestNewCoverageMonitor(t *testing.T) {
	m := NewCoverageMonitor(85.0)
	if m.threshold != 85.0 {
		t.Errorf("expected 85.0, got %f", m.threshold)
	}
}

func TestCoverageMonitor_ParseCoverageOutput(t *testing.T) {
	TestCoverageMonitor_ParseOutput(t)
}

func TestCoverageMonitor_ValidateThreshold(t *testing.T) {
	m := NewCoverageMonitor(85.0)
	data := map[string]float64{"pkg": 90.0}
	if _, ok := m.ValidateThreshold(data); !ok {
		t.Error("expected ok for 90.0 > 85.0")
	}
}

func TestCoverageMonitor_GenerateCoverageReport(t *testing.T) {
	m := NewCoverageMonitor(85.0)
	report := m.GenerateCoverageReport(map[string]float64{"pkg": 90.0})
	if !strings.Contains(report, "pkg") {
		t.Error("report missing package")
	}
}

// Property 18: CI Report Generation
// **Validates: Requirements 7.3**
func TestProperty_CoverageReportGeneration(t *testing.T) {
	f := func(packages map[string]float64) bool {
		monitor := NewCoverageMonitor(85.0)
		report := monitor.GenerateCoverageReport(packages)

		if report == "" {
			return false
		}

		// Check if all packages are in the report
		for pkg := range packages {
			if !strings.Contains(report, pkg) {
				return false
			}
		}
		return true
	}

	testutil.CheckProperty(t, f)
}
