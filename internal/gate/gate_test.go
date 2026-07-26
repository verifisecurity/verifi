package gate

import "testing"

// TestEvaluate walks every row of the MVP table: which evidence authorizes an
// apply-on-confirm, and which can only be proposed. One row per case, plus the
// residual-warning variants.
func TestEvaluate(t *testing.T) {
	cases := []struct {
		name        string
		ev          Evidence
		wantAuth    string
		wantWarning bool
	}{
		{
			name:     "remove unused direct dependency",
			ev:       Evidence{Action: "remove", Direct: true, Imported: false},
			wantAuth: Confirm,
		},
		{
			name:     "remove but code imports it",
			ev:       Evidence{Action: "remove", Direct: true, Imported: true},
			wantAuth: Propose,
		},
		{
			name:     "remove of a transitive dependency",
			ev:       Evidence{Action: "remove", Direct: false, Imported: false},
			wantAuth: Propose,
		},
		{
			name:     "upgrade a patch bump",
			ev:       Evidence{Action: "upgrade", TargetExists: true, Distance: "patch"},
			wantAuth: Confirm,
		},
		{
			name:     "upgrade a minor bump",
			ev:       Evidence{Action: "upgrade", TargetExists: true, Distance: "minor"},
			wantAuth: Confirm,
		},
		{
			name:        "upgrade a major bump warns but still confirms",
			ev:          Evidence{Action: "upgrade", TargetExists: true, Distance: "major"},
			wantAuth:    Confirm,
			wantWarning: true,
		},
		{
			name:     "upgrade with no published target",
			ev:       Evidence{Action: "upgrade", TargetExists: false, Distance: "minor"},
			wantAuth: Propose,
		},
		{
			name:     "no fix available",
			ev:       Evidence{Action: "none"},
			wantAuth: Propose,
		},
		{
			name:        "patch upgrade with a residual advisory warns",
			ev:          Evidence{Action: "upgrade", TargetExists: true, Distance: "patch", Residual: []string{"GHSA-xxxx"}},
			wantAuth:    Confirm,
			wantWarning: true,
		},
		{
			name:        "no fix with a residual advisory warns",
			ev:          Evidence{Action: "none", Residual: []string{"GHSA-yyyy"}},
			wantAuth:    Propose,
			wantWarning: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := Evaluate(tc.ev)
			if d.Authorization != tc.wantAuth {
				t.Errorf("authorization = %q, want %q", d.Authorization, tc.wantAuth)
			}
			if d.Rung != RungAdvisory {
				t.Errorf("rung = %q, want %q", d.Rung, RungAdvisory)
			}
			if len(d.Reasons) == 0 {
				t.Error("decision has no reasons; every row should explain itself")
			}
			if gotWarn := len(d.Warnings) > 0; gotWarn != tc.wantWarning {
				t.Errorf("has warning = %v, want %v (%v)", gotWarn, tc.wantWarning, d.Warnings)
			}
		})
	}
}

// TestUnattendedUnreachable is the release guard: no evidence may authorize an
// unattended apply. The gate can only ever confirm or propose.
func TestUnattendedUnreachable(t *testing.T) {
	actions := []string{"upgrade", "remove", "none", ""}
	distances := []string{"patch", "minor", "major", ""}
	bools := []bool{true, false}
	for _, a := range actions {
		for _, dist := range distances {
			for _, direct := range bools {
				for _, imported := range bools {
					for _, exists := range bools {
						d := Evaluate(Evidence{
							Action: a, Distance: dist, Direct: direct,
							Imported: imported, TargetExists: exists,
						})
						if d.Authorization == Unattended {
							t.Fatalf("gate returned unattended for %+v", Evidence{
								Action: a, Distance: dist, Direct: direct,
								Imported: imported, TargetExists: exists,
							})
						}
						if d.Authorization != Confirm && d.Authorization != Propose {
							t.Fatalf("authorization = %q, want confirm or propose", d.Authorization)
						}
					}
				}
			}
		}
	}
}
