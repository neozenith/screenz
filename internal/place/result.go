package place

import "github.com/neozenith/screenz/internal/mac"

// Result reports one window placement with the read-back frame — AX
// set-frame is a proposal that apps clamp silently, so the read back is the
// only honest success signal (ADR3.1). The type is unconstrained by build
// tags so the pure CLI pipeline compiles everywhere; Place itself is
// darwin-only.
type Result struct {
	Requested mac.CGRect `json:"requested"`
	Before    mac.CGRect `json:"before"`
	Actual    mac.CGRect `json:"actual"`
	Attempts  int        `json:"attempts"`
	OK        bool       `json:"ok"`
	Err       string     `json:"err,omitempty"`
}
