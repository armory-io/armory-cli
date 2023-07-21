// Package console
// The purpose of this package is to funnel all CLI console output, ensuring POSIX compliant use of stderr and stdout.
package console

import (
	"fmt"
	"github.com/samber/lo"
	"io"
	"os"
)

var (
	verbose           = lo.ToPtr(false)
	stdout  io.Writer = os.Stdout
	stderr  io.Writer = os.Stderr
)

type Options struct {
	Verbose *bool
	Stdout  io.Writer
	Stderr  io.Writer
}

func Configure(opts *Options) {
	if opts.Verbose != nil {
		verbose = opts.Verbose
	}

	if opts.Stdout != nil {
		stdout = opts.Stdout
	}

	if opts.Stderr != nil {
		stderr = opts.Stderr
	}
}

// Stdoutf formats according to a format specifier and writes the resulting string to stdout, swallowing any errors.
// DO NOT USE Stdoutf for debug statements, warnings, notifications, or errors.
// When an output format flag such as `-o json` is used, stdout should have a single parsable JSON object written to it, so that it can be piped and parsed by tools such as JQ.
func Stdoutf(format string, a ...any) {
	_, _ = fmt.Fprintf(stdout, format, a...)
}

// Stdoutln formats using the default formats for its operands and writes to stdout, swallowing any errors.
// Spaces are always added between operands and a newline is appended.
// DO NOT USE Stdoutln for debug statements, warnings, notifications, or errors.
// When an output format flag such as `-o json` is used, stdout should have a single parsable JSON object written to it, so that it can be piped and parsed by tools such as JQ.
func Stdoutln(a ...any) {
	_, _ = fmt.Fprintln(stdout, a...)
}

// Stderrf formats according to a format specifier and writes the resulting string to stderr, swallowing any errors.
// use Stderrf for output that should not be pipe-able to another command, IE any debug statements, warnings, notifications, or errors.
func Stderrf(format string, a ...any) {
	_, _ = fmt.Fprintf(stderr, format, a...)
}

// Stderrln formats using the default formats for its operands and writes to stderr, swallowing any errors.
// Spaces are always added between operands and a newline is appended.
// use Stderrln for output that should not be pipe-able to another command, IE any debug statements, warnings, notifications, or errors.
func Stderrln(a ...any) {
	_, _ = fmt.Fprintln(stderr, a...)
}

// Debugf formats according to a format specifier and writes the resulting string to stderr, swallowing any errors, if the enabled bool is true.
func Debugf(format string, a ...any) {
	if *verbose {
		_, _ = fmt.Fprintf(stderr, format, a...)
	}
}

// Debugln formats according to a format specifier and writes the resulting string to stderr, swallowing any errors, if the enabled bool is true.
// Spaces are always added between operands and a newline is appended.
func Debugln(a ...any) {
	if *verbose {
		_, _ = fmt.Fprintln(stderr, a...)
	}
}
