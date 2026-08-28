package profile

import (
	"path/filepath"
	"testing"
)

func TestDir(t *testing.T) {
	env := func(vars map[string]string) func(string) string {
		return func(k string) string { return vars[k] }
	}
	cases := []struct {
		name string
		vars map[string]string
		want string
	}{
		{"screenz home wins", map[string]string{"SCREENZ_HOME": "/dot/screenz", "XDG_CONFIG_HOME": "/dot/config"}, "/dot/screenz"},
		{"xdg second", map[string]string{"XDG_CONFIG_HOME": "/dot/config"}, filepath.Join("/dot/config", "screenz")},
		{"home fallback", map[string]string{}, filepath.Join("/Users/example", ".config", "screenz")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Dir(env(tc.vars), "/Users/example"); got != tc.want {
				t.Fatalf("Dir() = %q, want %q", got, tc.want)
			}
		})
	}
}
