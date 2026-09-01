package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// templateYAML is the commented starting point emitted by `screenz init`.
// Block style only — goccy loses comments inside flow style {…} (#608) —
// and every key mirrors the CLI rule grammar one-to-one (ADR4.1). Blank
// lines are kept here for readability but do not survive a later
// --save-profile (goccy/go-yaml#285).
const templateYAML = `# screenz profile "NAME"          -- emitted by: screenz init --profile NAME
# Apply with:   screenz apply --profile NAME
# Preview with: screenz apply --dry-run --profile NAME
version: 1
name: NAME

# Display aliases use the same terms as --display.
# An alias that does not resolve on this machine fails loudly (exit 1)
# before any window moves.
displays:
  dell-left:
    name: /LU28R55/ # regex against the display's localized name
    index: 1 # displays ordered top-to-bottom, then left-to-right
  dell-right:
    name: /LU28R55/
    index: 2
  laptop:
    built-in: true

# Rules run in order; a window is placed by the FIRST rule it matches.
rules:
  - match:
      bundle: com.microsoft.VSCode
    display: dell-left # alias, or an inline map such as {index: 1}
    region: maximize
  - match:
      app: Google Chrome
      title: /Work/ # regex; add i after the closing slash for case-insensitive
    display: dell-right
    region: left-half
    gap: 8 # points between window and region edge
  - match:
      bundle: com.microsoft.edgemac
    display: dell-right
    region: right-half
    tolerance: 5% # accept frames within 5% of the target size; default 0.5 (points)
    each: true # default; false == --first
    order: title # title | pid | existing
`

// Template returns the init template with the profile name substituted.
func Template(name string) string {
	return strings.ReplaceAll(templateYAML, "NAME", name)
}

// WriteTemplate writes the init template for a new profile. An existing
// file is refused unless force is set — enforced by O_EXCL at open, not a
// separate exists pre-check.
func WriteTemplate(path, name string, force bool) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	if !force {
		flags = os.O_WRONLY | os.O_CREATE | os.O_EXCL
	}
	f, err := os.OpenFile(path, flags, 0o644)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("%s already exists (use --force to overwrite)", path)
		}
		return err
	}
	defer f.Close()
	_, err = f.WriteString(Template(name))
	return err
}
