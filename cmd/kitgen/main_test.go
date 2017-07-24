package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestProcess(t *testing.T) {
	cases, err := filepath.Glob("testdata/*")
	if err != nil {
		t.Fatal(err)
	}

	testcase := func(dir string) {
		name := filepath.Base(dir)
		t.Run(name, func(t *testing.T) {
			inpath := filepath.Join(dir, "in.go")
			outpath := filepath.Join(dir, "out.go")

			in, err := ioutil.ReadFile(inpath)
			if err != nil {
				t.Fatal(inpath, err)
			}

			expected, err := ioutil.ReadFile(outpath)
			if err != nil {
				t.Fatal(outpath, err)
			}

			actualR, err := process(inpath, bytes.NewBuffer(in))
			if err != nil {
				t.Fatal(outpath, err)
			}

			actual, err := ioutil.ReadAll(actualR)
			if err != nil {
				t.Fatal(outpath, err)
			}

			if !bytes.Equal(expected, actual) {
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
		})
	}

	for _, dir := range cases {
		testcase(dir)
	}
}
