package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var (
	// Logf will always output.
	Logf = log.Printf

	// Debugf will only output if -debug is passed.
	Debugf = log.Printf
)

func main() {
	// Set up the flags and usage text.
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  %s [flags] <Go source files>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}
	var (
		debug = flag.Bool("debug", false, "Print debug information during parse")
		name  = flag.String("name", "Service", "Interface name to extract")
	)
	flag.Parse()
	if flag.NArg() <= 0 {
		fmt.Fprintf(os.Stderr, "Error: %s requires at least 1 Go source file as an argument.\n\n", os.Args[0])
		flag.Usage()
		os.Exit(1)
	}

	// Set up our output.
	log.SetFlags(0)
	if !*debug {
		Debugf = func(format string, args ...interface{}) {}
	}

	// Parse the files for the interface we will implement.
	imports, iface, err := parseFiles(*name, flag.Args()...)
	if err != nil {
		Logf("Parse failed: %v", err)
		os.Exit(1)
	}
	Logf(
		"Found interface %s, with %d method(s), requiring %d import(s)",
		iface.Name,
		len(iface.Methods),
		len(imports),
	)
	Logf("import (")
	for _, importStr := range imports {
		Logf("\t%s", importStr)
	}
	Logf(")")
	Logf("")
	Logf("%s", iface)

	// TODO(pb): request and response structs
	// TODO(pb): default interface implementation
	// TODO(pb): endpoint constructors
	// TODO(pb): default transport binding
}
