package term_test

import (
	"errors"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/term"
)

func ExampleNewLogger() {
	// Color errors red
	colorFn := func(keyvals ...interface{}) term.FgBgColor {
		for i := 1; i < len(keyvals); i += 2 {
			if _, ok := keyvals[i].(error); ok {
				return term.FgBgColor{Fg: term.White, Bg: term.Red}
			}
		}
		return term.FgBgColor{}
	}

	logger := term.NewLogger(os.Stdout, log.NewLogfmtLogger, colorFn)

	logger.Log("msg", "default color", "err", nil)
	logger.Log("msg", "colored because of error", "err", errors.New("coloring error"))
}
