package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/itchyny/gojq"
)

// jqOpts carries one command's --jq and --raw flag values. Filtering is the
// same contract on every command that emits JSON, so it is registered,
// validated and applied in one place (ADR-0027).
type jqOpts struct {
	query string
	raw   bool
}

// register binds both flags. Neither takes a one-letter alias: -j is --json,
// -r is apply's --region (ADR-0021), and -q reads as "quiet" everywhere else
// in Unix — a letter here would cost more surprise than it saves typing.
func (o *jqOpts) register(fs *flag.FlagSet) {
	fs.StringVar(&o.query, "jq", "", "filter the JSON output through this jq query")
	fs.BoolVar(&o.raw, "raw", false, "print string results from --jq unquoted")
}

// jqFilter is a compiled query plus the raw-string switch. A nil *jqFilter
// means no --jq was given, which is how emitJSON tells unfiltered output
// apart from a query that happens to be the identity.
type jqFilter struct {
	code *gojq.Code
	raw  bool
}

// resolve compiles the query before the command does any work, so a typo
// fails before a window moves. --jq implies --json, so neither has to be
// typed with the other; --raw alone has nothing to act on and says so rather
// than being quietly ignored (ADR2.2). A false second return is a usage
// error, which the caller reports as exit 2.
func (o *jqOpts) resolve(cmd string, jsonOut *bool, stderr io.Writer) (*jqFilter, bool) {
	if o.query == "" {
		if o.raw {
			fmt.Fprintf(stderr, "screenz %s: --raw needs a --jq query to filter\n", cmd)
			return nil, false
		}
		return nil, true
	}
	q, err := gojq.Parse(o.query)
	if err != nil {
		fmt.Fprintf(stderr, "screenz %s: --jq: %v\n", cmd, err)
		return nil, false
	}
	code, err := gojq.Compile(q)
	if err != nil {
		fmt.Fprintf(stderr, "screenz %s: --jq: %v\n", cmd, err)
		return nil, false
	}
	*jsonOut = true
	return &jqFilter{code: code, raw: o.raw}, true
}

// emitJSON writes v as indented JSON, or as the results of the --jq query
// when one was given. Filtered output matches what piping the unfiltered
// JSON into `jq -S` would have printed: one result per line, two-space
// indent, strings quoted unless --raw. The -S is not a choice — gojq holds
// objects in Go maps, so an emitted object's keys are always sorted rather
// than kept in schema order. It returns a process exit code: 1 when the
// query fails, 0 otherwise.
func emitJSON(stdout, stderr io.Writer, cmd string, v any, f *jqFilter) int {
	if f == nil {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(v)
		return 0
	}
	// gojq walks plain maps and slices, not our typed structs, so the report
	// is round-tripped through JSON to get there. UseNumber carries every
	// number across as the digits encoding/json just wrote, rather than as
	// a float64 that has to be formatted back — so a filtered frame or
	// window id reads exactly as the unfiltered report spells it. Key order
	// is still lost in the map; only the numbers survive the trip intact.
	b, err := json.Marshal(v)
	if err != nil {
		fmt.Fprintf(stderr, "screenz %s: --jq: %v\n", cmd, err)
		return 1
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	var doc any
	// Cannot fail: these are bytes encoding/json produced a statement ago.
	_ = dec.Decode(&doc)

	iter := f.code.Run(doc)
	for {
		res, ok := iter.Next()
		if !ok {
			return 0
		}
		if err, ok := res.(error); ok {
			// `halt` is a clean stop, not a failure; `halt_error` carries a
			// value and is reported like any other runtime error.
			var halt *gojq.HaltError
			if errors.As(err, &halt) && halt.Value() == nil {
				return 0
			}
			fmt.Fprintf(stderr, "screenz %s: --jq: %v\n", cmd, err)
			return 1
		}
		if s, ok := res.(string); ok && f.raw {
			fmt.Fprintln(stdout, s)
			continue
		}
		// jq's own encoding (its escaping rules differ from encoding/json),
		// then jq's own two-space indent. Marshal only ever sees the types a
		// gojq iterator emits, and Indent only bytes Marshal just produced,
		// so neither can fail.
		out, _ := gojq.Marshal(res)
		var pretty bytes.Buffer
		_ = json.Indent(&pretty, out, "", "  ")
		fmt.Fprintln(stdout, pretty.String())
	}
}
