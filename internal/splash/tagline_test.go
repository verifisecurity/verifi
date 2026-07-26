package splash

import "testing"

// TestTagline pins the thing that went stale: the banner states the build it is,
// not a promise about what is coming. A promise outlives its truth.
func TestTagline(t *testing.T) {
	tests := []struct{ version, want string }{
		{"0.1.1", "v0.1.1 · pre-1.0"},
		{"v0.2.0", "v0.2.0 · pre-1.0"},
		{"dev", "dev build · pre-1.0"},
		{"", "dev build · pre-1.0"},
		{"  1.0.0  ", "v1.0.0 · pre-1.0"},
	}
	orig := Version
	defer func() { Version = orig }()

	for _, tt := range tests {
		Version = tt.version
		if got := tagline(); got != tt.want {
			t.Errorf("Version=%q: tagline() = %q, want %q", tt.version, got, tt.want)
		}
	}
}

// TestTaglineMakesNoPromise guards the specific regression: no "coming soon" on
// a tool that has shipped.
func TestTaglineMakesNoPromise(t *testing.T) {
	orig := Version
	defer func() { Version = orig }()
	for _, v := range []string{"dev", "0.1.1", "2.0.0"} {
		Version = v
		if got := tagline(); got == "coming soon · beta" {
			t.Errorf("Version=%q still advertises coming soon", v)
		}
	}
}
