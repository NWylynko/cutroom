package cli

import (
	"fmt"
	"strings"

	"github.com/cutroom/cutroom/internal/deps"
)

func init() {
	Register(&Command{
		Name:     "doctor",
		Category: "start",
		Summary:  "Check that ffmpeg, whisper.cpp and a speech model are installed; print fixes",
		Run: func(c *Ctx) (*Result, error) {
			all := deps.All()
			var b strings.Builder
			bad := 0
			for _, s := range all {
				mark := "ok  "
				if !s.OK {
					mark, bad = "FAIL", bad+1
				}
				detail := s.Path
				if s.Version != "" {
					detail = s.Version + "  " + s.Path
				}
				if !s.OK {
					detail = s.Problem
				}
				fmt.Fprintf(&b, "%s %-14s %s\n", mark, s.Name, detail)
				if !s.OK && s.Fix != "" {
					fmt.Fprintf(&b, "     fix: %s\n", s.Fix)
				}
			}
			res := &Result{Data: map[string]any{"ok": bad == 0, "dependencies": all}}
			if bad > 0 {
				fmt.Fprintf(&b, "%d problem(s). Run each fix, then `cutroom doctor` again.", bad)
				res.Exit = ExitDep
			} else {
				b.WriteString("all dependencies ok")
				res.Next = []string{"cutroom init <video file>"}
			}
			res.Text = b.String()
			return res, nil
		},
	})
}
