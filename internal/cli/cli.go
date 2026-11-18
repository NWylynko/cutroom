// Package cli is cutroom's command layer. Everything here is shaped for an AI agent reader:
// compact text by default, complete JSON with --json, a "next:" hint on every success, and
// errors that say how to fix themselves (SPEC.md, design rules 2–4).
package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

// Version is set at build time with -ldflags "-X github.com/cutroom/cutroom/internal/cli.Version=…".
var Version = "0.0.0-dev"

// Exit codes (SPEC.md rule 4).
const (
	ExitOK        = 0
	ExitUsage     = 1
	ExitInvalid   = 2
	ExitDep       = 3
	ExitCheckFail = 4
	ExitExternal  = 5
)

// Error is the only error type commands should return. Code is a stable kebab-case id agents can
// match on; Fix is the exact next action.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Fix     string `json:"fix,omitempty"`
	Exit    int    `json:"-"`
}

func (e *Error) Error() string { return e.Message }

func Errf(exit int, code, fix, format string, a ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, a...), Fix: fix, Exit: exit}
}

// Result is what a command produces. Text is the compact human/agent view; Data is the full
// structured view for --json. Next lists 1–3 suggested commands.
type Result struct {
	Text string
	Data any
	Next []string
	Exit int // non-zero for "ran fine but the answer is a failure" (e.g. check failed)
}

// Flag describes one option. Kind is "bool", "string", "int" or "float".
type Flag struct {
	Name    string
	Kind    string
	Default string
	Help    string
	Many    bool // may repeat; values collected in order
}

// Command is one subcommand.
type Command struct {
	Name     string
	Args     string // usage of positional args, e.g. "<time|range>"
	Summary  string // one line
	Help     string // longer text; examples encouraged
	Flags    []Flag
	MinArgs  int
	MaxArgs  int // -1 = unlimited
	Run      func(ctx *Ctx) (*Result, error)
	Hidden   bool
	Category string
}

// Ctx is passed to Run.
type Ctx struct {
	Cmd   *Command
	Args  []string
	vals  map[string][]string
	JSON  bool
	Dir   string // project dir (--project or cwd)
	Out   io.Writer
	ErrW  io.Writer
	Stdin io.Reader
}

func (c *Ctx) Str(name string) string {
	if v := c.vals[name]; len(v) > 0 {
		return v[len(v)-1]
	}
	return c.flag(name).Default
}

func (c *Ctx) Strs(name string) []string { return c.vals[name] }

func (c *Ctx) Has(name string) bool { return len(c.vals[name]) > 0 }

func (c *Ctx) Bool(name string) bool { return c.Str(name) == "true" }

func (c *Ctx) Int(name string) int { n, _ := strconv.Atoi(c.Str(name)); return n }

func (c *Ctx) Float(name string) float64 { f, _ := strconv.ParseFloat(c.Str(name), 64); return f }

func (c *Ctx) flag(name string) Flag {
	for _, f := range append(c.Cmd.Flags, globalFlags...) {
		if f.Name == name {
			return f
		}
	}
	return Flag{Name: name} // undeclared: behaves as unset (TestEveryCommandRunsWithoutPanic guards this)
}

var globalFlags = []Flag{
	{Name: "json", Kind: "bool", Help: "Print the complete result as one JSON object"},
	{Name: "project", Kind: "string", Default: ".", Help: "Project directory (default: current directory)"},
	{Name: "help", Kind: "bool", Help: "Show help for this command"},
}

var registry = map[string]*Command{}

// Register adds a command; called from init() in each command file.
func Register(c *Command) {
	if c.MaxArgs == 0 && c.Args != "" {
		c.MaxArgs = -1
	}
	registry[c.Name] = c
}

// Commands returns visible commands sorted by category then name.
func Commands() []*Command {
	var out []*Command
	for _, c := range registry {
		if !c.Hidden {
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if catOrder(out[i].Category) != catOrder(out[j].Category) {
			return catOrder(out[i].Category) < catOrder(out[j].Category)
		}
		return out[i].Name < out[j].Name
	})
	return out
}

var categories = []string{"start", "perceive", "edit", "render", "verify", "deliver", "other"}

func catOrder(c string) int {
	for i, x := range categories {
		if x == c {
			return i
		}
	}
	return len(categories)
}

// Run executes argv (without the program name) and returns the exit code.
func Run(argv []string, stdout, stderr io.Writer, stdin io.Reader) int {
	jsonMode := false
	for _, a := range argv {
		if a == "--json" || a == "--json=true" {
			jsonMode = true
		}
	}
	res, err := dispatch(argv, stdout, stderr, stdin, jsonMode)
	if err != nil {
		e, ok := err.(*Error)
		if !ok {
			e = &Error{Code: "internal", Message: err.Error(), Exit: ExitExternal,
				Fix: "this is a cutroom bug or an unexpected tool failure; rerun with the same args, and report it if it repeats"}
		}
		if jsonMode {
			writeJSON(stdout, map[string]any{"ok": false, "error": e})
		} else {
			fmt.Fprintf(stderr, "error[%s]: %s\n", e.Code, e.Message)
			if e.Fix != "" {
				fmt.Fprintf(stderr, "fix: %s\n", e.Fix)
			}
		}
		return e.Exit
	}
	if jsonMode {
		obj := map[string]any{"ok": res.Exit == 0}
		if res.Data != nil {
			obj["result"] = res.Data
		}
		if len(res.Next) > 0 {
			obj["next"] = res.Next
		}
		writeJSON(stdout, obj)
	} else {
		text := strings.TrimRight(res.Text, "\n")
		if text != "" {
			fmt.Fprintln(stdout, text)
		}
		for _, n := range res.Next {
			fmt.Fprintf(stdout, "next: %s\n", n)
		}
	}
	return res.Exit
}

func writeJSON(w io.Writer, v any) {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func dispatch(argv []string, stdout, stderr io.Writer, stdin io.Reader, jsonMode bool) (res *Result, err error) {
	defer func() {
		if r := recover(); r != nil {
			res, err = nil, &Error{Code: "internal", Exit: ExitExternal, Message: fmt.Sprintf("cutroom crashed: %v", r),
				Fix: "this is a cutroom bug; try a different approach to the same step, and report the command"}
		}
	}()
	if len(argv) == 0 || argv[0] == "--help" || argv[0] == "-h" || argv[0] == "help" {
		if len(argv) > 1 && argv[0] == "help" {
			if c, ok := registry[argv[1]]; ok {
				return &Result{Text: commandHelp(c)}, nil
			}
		}
		return &Result{Text: overview()}, nil
	}
	if argv[0] == "--version" || argv[0] == "version" {
		return &Result{Text: "cutroom " + Version, Data: map[string]string{"version": Version}}, nil
	}
	name := argv[0]
	cmd, ok := registry[name]
	if !ok {
		fix := "run `cutroom` to list commands"
		if s := suggest(name, commandNames()); s != "" {
			fix = fmt.Sprintf("did you mean `cutroom %s`?", s)
		}
		return nil, Errf(ExitUsage, "unknown-command", fix, "unknown command %q", name)
	}
	ctx := &Ctx{Cmd: cmd, vals: map[string][]string{}, JSON: jsonMode, Out: stdout, ErrW: stderr, Stdin: stdin}
	if err := parseArgs(ctx, argv[1:]); err != nil {
		return nil, err
	}
	if ctx.Bool("help") {
		return &Result{Text: commandHelp(cmd)}, nil
	}
	ctx.Dir = ctx.Str("project")
	if len(ctx.Args) < cmd.MinArgs || (cmd.MaxArgs >= 0 && len(ctx.Args) > cmd.MaxArgs) {
		return nil, Errf(ExitUsage, "bad-args", "usage: "+usageLine(cmd),
			"%s expects %s, got %d argument(s)", cmd.Name, argCount(cmd), len(ctx.Args))
	}
	res, err = cmd.Run(ctx)
	if err != nil {
		return nil, err
	}
	if res == nil {
		res = &Result{}
	}
	return res, nil
}

func argCount(c *Command) string {
	switch {
	case c.MaxArgs == c.MinArgs:
		return strconv.Itoa(c.MinArgs)
	case c.MaxArgs < 0:
		return fmt.Sprintf("at least %d", c.MinArgs)
	default:
		return fmt.Sprintf("%d to %d", c.MinArgs, c.MaxArgs)
	}
}

func commandNames() []string {
	var n []string
	for k, c := range registry {
		if !c.Hidden {
			n = append(n, k)
		}
	}
	return n
}

func parseArgs(ctx *Ctx, args []string) error {
	flags := map[string]Flag{}
	for _, f := range append(append([]Flag{}, ctx.Cmd.Flags...), globalFlags...) {
		flags[f.Name] = f
	}
	names := func() []string {
		var n []string
		for k := range flags {
			n = append(n, k)
		}
		return n
	}
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			ctx.Args = append(ctx.Args, args[i+1:]...)
			break
		}
		if !strings.HasPrefix(a, "--") || len(a) == 2 {
			if strings.HasPrefix(a, "-") && len(a) > 1 && !isNumberish(a) {
				return Errf(ExitUsage, "bad-flag", fmt.Sprintf("use the long form, e.g. --%s", strings.TrimLeft(a, "-")),
					"short flags are not supported: %s", a)
			}
			ctx.Args = append(ctx.Args, a)
			continue
		}
		name, val, hasVal := strings.Cut(a[2:], "=")
		f, ok := flags[name]
		if !ok {
			fix := fmt.Sprintf("run `cutroom %s --help` to list flags", ctx.Cmd.Name)
			if s := suggest(name, names()); s != "" {
				fix = fmt.Sprintf("did you mean --%s? (`cutroom %s --help` lists all flags)", s, ctx.Cmd.Name)
			}
			return Errf(ExitUsage, "unknown-flag", fix, "%s has no flag --%s", ctx.Cmd.Name, name)
		}
		if f.Kind == "bool" {
			if !hasVal {
				val = "true"
			}
			if val != "true" && val != "false" {
				return Errf(ExitUsage, "bad-flag-value", fmt.Sprintf("use --%s or --%s=false", name, name), "--%s is a switch, got %q", name, val)
			}
		} else if !hasVal {
			if i+1 >= len(args) {
				return Errf(ExitUsage, "missing-flag-value", fmt.Sprintf("give a value: --%s <%s>", name, f.Kind), "--%s needs a value", name)
			}
			i++
			val = args[i]
		}
		switch f.Kind {
		case "int":
			if _, err := strconv.Atoi(val); err != nil {
				return Errf(ExitUsage, "bad-flag-value", fmt.Sprintf("--%s takes a whole number", name), "--%s: %q is not an integer", name, val)
			}
		case "float":
			if _, err := strconv.ParseFloat(val, 64); err != nil {
				return Errf(ExitUsage, "bad-flag-value", fmt.Sprintf("--%s takes a number like 0.4", name), "--%s: %q is not a number", name, val)
			}
		}
		if !f.Many {
			ctx.vals[name] = nil
		}
		ctx.vals[name] = append(ctx.vals[name], val)
	}
	return nil
}

func isNumberish(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// suggest returns the closest candidate within a small edit distance, or "".
func suggest(word string, candidates []string) string {
	best, bestD := "", 1<<30
	for _, c := range candidates {
		d := levenshtein(strings.ToLower(word), strings.ToLower(c))
		if strings.HasPrefix(c, word) || strings.HasPrefix(word, c) {
			d = min(d, 1)
		}
		if d < bestD || (d == bestD && c < best) {
			best, bestD = c, d
		}
	}
	limit := 2
	if len(word) > 6 {
		limit = 3
	}
	if bestD <= limit {
		return best
	}
	return ""
}

func levenshtein(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	prev := make([]int, len(rb)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ra); i++ {
		cur := make([]int, len(rb)+1)
		cur[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev = cur
	}
	return prev[len(rb)]
}

func usageLine(c *Command) string {
	s := "cutroom " + c.Name
	if c.Args != "" {
		s += " " + c.Args
	}
	if len(c.Flags) > 0 {
		s += " [flags]"
	}
	return s
}

func commandHelp(c *Command) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n  %s\n", usageLine(c), c.Summary)
	if c.Help != "" {
		fmt.Fprintf(&b, "\n%s\n", strings.TrimRight(c.Help, "\n"))
	}
	if len(c.Flags) > 0 {
		b.WriteString("\nflags:\n")
		for _, f := range c.Flags {
			arg := ""
			if f.Kind != "bool" {
				arg = " <" + f.Kind + ">"
			}
			def := ""
			if f.Default != "" && f.Kind != "bool" {
				def = " (default " + f.Default + ")"
			}
			rep := ""
			if f.Many {
				rep = " (repeatable)"
			}
			fmt.Fprintf(&b, "  --%s%s  %s%s%s\n", f.Name, arg, f.Help, def, rep)
		}
	}
	b.WriteString("\nglobal: --json (full result as JSON), --project <dir>")
	return b.String()
}

func overview() string {
	var b strings.Builder
	fmt.Fprintf(&b, "cutroom %s — video editing for AI agents (ffmpeg underneath)\n", Version)
	b.WriteString("New here? Run `cutroom guide` first: it is the manual for this version.\n")
	cat := ""
	for _, c := range Commands() {
		if c.Category != cat {
			cat = c.Category
			fmt.Fprintf(&b, "\n%s:\n", cat)
		}
		fmt.Fprintf(&b, "  %-11s %s\n", c.Name, c.Summary)
	}
	b.WriteString("\nEvery command: --help, --json (complete result), --project <dir>.")
	return b.String()
}
