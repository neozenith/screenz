package demo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neozenith/screenz/internal/layout"
	"github.com/neozenith/screenz/internal/mac"
)

const worldJSON = `{
  "schema": 1,
  "displays": [
    {"index": 1, "name": "Built-in Retina Display", "main": true, "built_in": true},
    {"index": 2, "name": "LU28R55 (1)"}
  ],
  "windows": [
    {"pid": 1, "id": 100, "app": "Code", "bundle": "com.microsoft.VSCode", "title": "Welcome"}
  ]
}`

func writeWorld(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "world.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadReplaysWorld(t *testing.T) {
	snap, err := Load(writeWorld(t, worldJSON))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(snap.Displays) != 2 || snap.Displays[1].Name != "LU28R55 (1)" {
		t.Fatalf("displays not replayed: %+v", snap.Displays)
	}
	if len(snap.Windows) != 1 || snap.Windows[0].Bundle != "com.microsoft.VSCode" {
		t.Fatalf("windows not replayed: %+v", snap.Windows)
	}
	if el := snap.AppEl(1); el != (mac.AXElement{}) {
		t.Fatalf("replayed snapshot must carry zero AX elements, got %+v", el)
	}
}

func TestLoadMissingFile(t *testing.T) {
	if _, err := Load(filepath.Join(t.TempDir(), "absent.json")); err == nil {
		t.Fatal("want error for missing file")
	}
}

func TestLoadBadJSON(t *testing.T) {
	_, err := Load(writeWorld(t, "{not json"))
	if err == nil || !strings.Contains(err.Error(), "world.json") {
		t.Fatalf("want parse error naming the file, got %v", err)
	}
}

func TestLoadWrongSchema(t *testing.T) {
	_, err := Load(writeWorld(t, `{"schema": 2}`))
	if err == nil || !strings.Contains(err.Error(), "unsupported world schema 2") {
		t.Fatalf("want schema error, got %v", err)
	}
}

func TestPlaceSimulatesTargetAsReadBack(t *testing.T) {
	target := mac.CGRect{Origin: mac.CGPoint{X: 10, Y: 20}, Size: mac.CGSize{W: 300, H: 400}}
	res := Place(mac.AXElement{}, mac.AXElement{}, target, layout.Tolerance{})
	if !res.OK || res.Actual != target || res.Requested != target || res.Attempts != 1 || res.Err != "" {
		t.Fatalf("unexpected simulated result: %+v", res)
	}
}
