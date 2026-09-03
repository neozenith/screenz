package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/neozenith/screenz/internal/discover"
)

// TestJQFiltersEveryJSONCommand pins the contract that --jq is available
// wherever --json is, and that it implies --json rather than needing it.
func TestJQFiltersEveryJSONCommand(t *testing.T) {
	profiled, home := profDeps(t)
	writeProfile(t, home, "office", officeProfileYAML)

	for _, tc := range []struct {
		name string
		args []string
		deps Deps
		want string
	}{
		{"status", []string{"status", "--jq", ".windows | length"}, snapDeps(officeSnapshot(), nil), "3\n"},
		{"doctor", []string{"doctor", "--jq", ".accessibility"}, deps(officeSys(true)), "\"trusted\"\n"},
		{"list", []string{"list", "--jq", "[.profiles[].name]"}, profiled, "[\n  \"office\"\n]\n"},
		{"apply", []string{"apply", "--jq", ".actions | length", "--match", "bundle=com.microsoft.VSCode", "--display", "index=2"}, applyDeps("ok"), "1\n"},
		{"apply-dry-run", []string{"apply", "--dry-run", "--jq", ".plan.actions | length", "--match", "bundle=com.microsoft.VSCode", "--display", "index=2"}, applyDeps("ok"), "1\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			code, out, errOut := run(t, tc.args, tc.deps)
			if code != 0 {
				t.Fatalf("exit = %d; stderr: %s", code, errOut)
			}
			if out != tc.want {
				t.Errorf("stdout = %q, want %q", out, tc.want)
			}
		})
	}
}

// TestJQRawUnquotesStrings covers the one deviation from jq's defaults the
// CLI had to make: jq spells this -r, which on apply is already --region.
func TestJQRawUnquotesStrings(t *testing.T) {
	code, out, errOut := run(t, []string{"status", "--jq", ".windows[].app", "--raw"}, snapDeps(officeSnapshot(), nil))
	if code != 0 {
		t.Fatalf("exit = %d; stderr: %s", code, errOut)
	}
	if want := "Code\nCode\nGoogle Chrome\n"; out != want {
		t.Errorf("stdout = %q, want %q", out, want)
	}
	// Without --raw the same query keeps jq's quoting.
	_, quoted, _ := run(t, []string{"status", "--jq", ".windows[].app"}, snapDeps(officeSnapshot(), nil))
	if !strings.Contains(quoted, `"Google Chrome"`) {
		t.Errorf("default output dropped jq's quoting: %q", quoted)
	}
}

// TestJQRawOnlyIsUsageError: --raw alone has nothing to act on. Ignoring it
// would let a script think it was getting unquoted strings (ADR2.2).
func TestJQRawOnlyIsUsageError(t *testing.T) {
	code, out, errOut := run(t, []string{"status", "--raw"}, snapDeps(officeSnapshot(), nil))
	if code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
	if out != "" {
		t.Errorf("stdout not empty: %q", out)
	}
	if !strings.Contains(errOut, "--raw needs a --jq query") {
		t.Errorf("stderr = %q", errOut)
	}
}

// TestJQBadQueryIsUsageError covers both compile-time rejections, and pins
// them to exit 2 — a query that cannot run is a flag value the user got
// wrong, not a failure of the machine.
func TestJQBadQueryIsUsageError(t *testing.T) {
	for _, tc := range []struct{ name, query, want string }{
		{"syntax", ".windows[", "unexpected EOF"},
		{"undefined function", "nosuchfn", "function not defined"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			code, out, errOut := run(t, []string{"status", "--jq", tc.query}, snapDeps(officeSnapshot(), nil))
			if code != 2 {
				t.Fatalf("exit = %d, want 2", code)
			}
			if out != "" {
				t.Errorf("stdout not empty: %q", out)
			}
			if !strings.Contains(errOut, tc.want) {
				t.Errorf("stderr = %q, want it to mention %q", errOut, tc.want)
			}
		})
	}
}

// TestJQBadQueryRejectedByEveryCommand: each command compiles the query
// itself, so each needs proving — and each must refuse before it reads the
// world, saves a profile or moves a window.
func TestJQBadQueryRejectedByEveryCommand(t *testing.T) {
	profiled, home := profDeps(t)
	writeProfile(t, home, "office", officeProfileYAML)

	for _, tc := range []struct {
		name string
		args []string
		deps Deps
	}{
		{"doctor", []string{"doctor", "--jq", ".windows["}, deps(officeSys(true))},
		{"list", []string{"list", "--jq", ".windows["}, profiled},
		{"apply", []string{"apply", "--jq", ".windows[", "--match", "bundle=com.microsoft.VSCode", "--display", "index=2"}, applyDeps("ok")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			code, out, errOut := run(t, tc.args, tc.deps)
			if code != 2 {
				t.Fatalf("exit = %d, want 2", code)
			}
			if out != "" {
				t.Errorf("stdout not empty: %q", out)
			}
			if !strings.Contains(errOut, "--jq:") {
				t.Errorf("stderr = %q, want it to name --jq", errOut)
			}
		})
	}
}

// TestJQCompilesBeforeTheWorldIsRead is the reason resolve runs where it
// does: a mistyped query must not cost a snapshot, let alone a placement.
func TestJQCompilesBeforeTheWorldIsRead(t *testing.T) {
	d := snapDeps(officeSnapshot(), nil)
	d.Snapshot = func() (discover.Snapshot, error) {
		t.Error("snapshot taken despite an uncompilable --jq query")
		return officeSnapshot(), nil
	}
	if code, _, _ := run(t, []string{"status", "--jq", ".windows["}, d); code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
}

// TestJQRuntimeErrorExits1 separates a query that would not compile (the
// user's typo, exit 2) from one that compiled and then met data it could
// not handle (exit 1).
func TestJQRuntimeErrorExits1(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		deps Deps
	}{
		{"status", []string{"status", "--jq", ".windows | .nope"}, snapDeps(officeSnapshot(), nil)},
		{"doctor", []string{"doctor", "--jq", ".displays | .nope"}, deps(officeSys(true))},
		{"apply", []string{"apply", "--jq", ".actions | .nope", "--match", "bundle=com.microsoft.VSCode", "--display", "index=2"}, applyDeps("ok")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			code, _, errOut := run(t, tc.args, tc.deps)
			if code != 1 {
				t.Fatalf("exit = %d, want 1", code)
			}
			if !strings.Contains(errOut, "--jq:") {
				t.Errorf("stderr = %q, want it to name --jq", errOut)
			}
		})
	}
}

// TestJQHaltStopsCleanly: jq's `halt` ends the stream, it does not fail.
func TestJQHaltStopsCleanly(t *testing.T) {
	code, out, errOut := run(t, []string{"status", "--jq", ".windows[0].app, halt", "--raw"}, snapDeps(officeSnapshot(), nil))
	if code != 0 {
		t.Fatalf("exit = %d, want 0; stderr: %s", code, errOut)
	}
	if out != "Code\n" {
		t.Errorf("stdout = %q, want %q", out, "Code\n")
	}
}

// TestEmitJSONUnmarshalableValue reaches the one branch no command can:
// every report the CLI builds is JSON-encodable by construction, so the
// guard is exercised directly rather than left unproven.
func TestEmitJSONUnmarshalableValue(t *testing.T) {
	filter, ok := (&jqOpts{query: "."}).resolve("status", new(bool), &bytes.Buffer{})
	if !ok {
		t.Fatal("identity query did not compile")
	}
	var stdout, stderr bytes.Buffer
	if code := emitJSON(&stdout, &stderr, "status", make(chan int), filter); code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout not empty: %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "unsupported type") {
		t.Errorf("stderr = %q", stderr.String())
	}
}
