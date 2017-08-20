package main

import (
	"bytes"
	"flag"
	"io"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "update golden files")

func TestProcess(t *testing.T) {
	cases, err := filepath.Glob("testdata/*")
	if err != nil {
		t.Fatal(err)
	}

	laidout := func(t *testing.T, inpath, dir, kind string, layout layout, in []byte) {
		t.Run(kind, func(t *testing.T) {
			tree, err := process(inpath, bytes.NewBuffer(in), layout)
			if err != nil {
				t.Fatal(inpath, err)
			}

			if *update {
				err := splat(filepath.Join(dir, kind), tree)
				if err != nil {
					t.Fatal(kind, err)
				}
			}

			for fn, buf := range tree {
				actual, err := ioutil.ReadAll(buf)
				if err != nil {
					t.Fatal(kind, fn, err)
				}

				outpath := filepath.Join(dir, kind, fn)

				expected, err := ioutil.ReadFile(outpath)
				if err != nil {
					t.Fatal(outpath, err)
				}

				if !bytes.Equal(expected, actual) {
					name := kind + fn
					errfile, err := ioutil.TempFile("", name)
					if err != nil {
						t.Fatal("opening tempfile for output", err)
					}
					io.WriteString(errfile, string(actual))

					diffCmd := exec.Command("diff", outpath, errfile.Name())
					diffOut, _ := diffCmd.Output()
					t.Log(string(diffOut))
					t.Errorf("Processing output didn't match %q. Results recorded in %q.", outpath, errfile.Name())
				}
			}
		})
	}

	testcase := func(dir string) {
		name := filepath.Base(dir)
		t.Run(name, func(t *testing.T) {
			inpath := filepath.Join(dir, "in.go")

			in, err := ioutil.ReadFile(inpath)
			if err != nil {
				t.Fatal(inpath, err)
			}
			laidout(t, inpath, dir, "flat", flat{}, in)
			laidout(t, inpath, dir, "default", deflayout{}, in)
		})
	}

	for _, dir := range cases {
		testcase(dir)
	}
}
