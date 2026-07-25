package fix

import (
	"reflect"
	"testing"

	"github.com/verifisecurity/verifi/internal/reason"
)

func TestBuild(t *testing.T) {
	recs := []reason.Recommendation{
		{Name: "lodash", Action: "upgrade", Current: "4.17.11", Target: "4.18.0", Reason: "clears X"},
		{Name: "leftpad", Action: "none"}, // no fix, should be skipped
	}
	p := Build(recs)

	if len(p.Actions) != 1 {
		t.Fatalf("actions = %d, want 1", len(p.Actions))
	}
	a := p.Actions[0]
	if a.Kind != "upgrade" || a.Name != "lodash" || a.From != "4.17.11" || a.To != "4.18.0" {
		t.Errorf("action = %+v", a)
	}
	want := []string{"npm", "install", "lodash@4.18.0"}
	if !reflect.DeepEqual(a.Command, want) {
		t.Errorf("command = %v, want %v", a.Command, want)
	}
	if len(p.Skipped) != 1 || p.Skipped[0] != "leftpad" {
		t.Errorf("skipped = %v, want [leftpad]", p.Skipped)
	}
	if p.Empty() {
		t.Error("plan should not be empty")
	}
}

func TestBuild_Empty(t *testing.T) {
	if !Build(nil).Empty() {
		t.Error("nil recommendations should give an empty plan")
	}
}
