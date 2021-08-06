package helpers

import (
	"fmt"
	"github.com/armory/armory-cli/internal/deng/protobuff"
	"io"
	"os"
	"strconv"
	"strings"
)

// ANSI escape codes
const (
	escape    = "\x1b"
	NoFormat  = 0
	Bold      = 1
	FgBlack   = 30
	FgRed     = 31
	FgGreen   = 32
	FgYellow  = 33
	FgBlue    = 34
	FgMagenta = 35
	FgCyan    = 36
	FgWhite   = 37
	FgDefault = 39
	FgHiBlue  = 94
)

// Clear clears the terminal for updates for live watching of objects
func Clear(out io.Writer) {
	fmt.Fprint(out, "\033[H\033[2J")
	fmt.Fprint(out, "\033[0;0H")
}

func AnsiFormat(s string, codes ...int) string {
	if os.Getenv("TERM") == "dumb" || len(codes) == 0 {
		return s
	}
	codeStrs := make([]string, len(codes))
	for i, code := range codes {
		codeStrs[i] = strconv.Itoa(code)
	}
	sequence := strings.Join(codeStrs, ";")
	return fmt.Sprintf("%s[%sm%s%s[%dm", escape, sequence, s, escape, NoFormat)
}

func Status(status protobuff.Status) string {
	switch status {
	case protobuff.Status_SUCCEEDED:
		return AnsiFormat("Success", FgGreen)
	case protobuff.Status_PENDING:
		return AnsiFormat("Pending", FgBlue)
	case protobuff.Status_RESOLVED:
		return AnsiFormat("Resolved", FgCyan)
	case protobuff.Status_FAILED:
		return AnsiFormat("Failed", FgRed)
	case protobuff.Status_FAILED_CLEANING:
		return AnsiFormat("Failed, cleaning", FgRed)
	case protobuff.Status_ABORTED:
		return AnsiFormat("Aborted", FgRed)
	case protobuff.Status_PAUSED:
		return AnsiFormat("Paused", FgYellow)
	case protobuff.Status_QUEUED:
		return AnsiFormat("Queued", FgCyan)
	case protobuff.Status_SUCCEEDED_CLEANING:
		return AnsiFormat("Success, cleaning", FgGreen)
	}
	return status.String()
}
