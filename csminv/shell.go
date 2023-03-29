/*
MIT License

(C) Copyright 2022 Hewlett Packard Enterprise Development LP

Permission is hereby granted, free of charge, to any person obtaining a
copy of this software and associated documentation files (the "Software"),
to deal in the Software without restriction, including without limitation
the rights to use, copy, modify, merge, publish, distribute, sublicense,
and/or sell copies of the Software, and to permit persons to whom the
Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included
in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
OTHER DEALINGS IN THE SOFTWARE.
*/

package csminv

import (
	"fmt"
	_ "image/png"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
)

// checkCommandExists checks if a command exists in the $PATH
func checkCommandExists(cmd string) {
	path, err := exec.LookPath(cmd)
	if err != nil {
		fmt.Printf("'%s' executable not found in $PATH\n", path)
	}
}

// copyAndCapture copies the output of a command to stdout and returns the output
// https://www.codeeval.dev/edit?gistid=7cca5f1526b2785e03febc16542b96ea
func copyAndCapture(w io.Writer, r io.Reader) ([]byte, error) {
	var out []byte
	buf := make([]byte, 1024, 1024)
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			d := buf[:n]
			out = append(out, d...)
			_, err := w.Write(d)
			if err != nil {
				return out, err
			}
		}
		if err != nil {
			// Read returns io.EOF at the end of file, which is not an error for us
			if err == io.EOF {
				err = nil
			}
			return out, err
		}
	}
}

// shell runs shell commands and any arguments
// https://gist.github.com/hivefans/ffeaf3964924c943dd7ed83b406bbdea?permalink_comment_id=3097813#gistcomment-3097813
func shell(command string, args []string) (stdout []byte, stderr []byte, err error) {
	cmd := exec.Command(command, args...)
	var errStdout, errStderr error
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()
	err = cmd.Start()
	if err != nil {
		log.Fatalf("cmd.Start() failed with '%s'\n", err)
		return nil, nil, err
	}
	// cmd.Wait() should be called only after we finish reading
	// from stdoutIn and stderrIn.
	// wg ensures that we finish
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		stdout, errStdout = copyAndCapture(os.Stdout, stdoutIn)
		wg.Done()
	}()

	stderr, errStderr = copyAndCapture(os.Stderr, stderrIn)
	wg.Wait()

	err = cmd.Wait()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	if errStdout != nil || errStderr != nil {
		log.Fatal("failed to capture stdout or stderr\n")
	}
	if stderr != nil {
		return nil, stderr, err
	}
	if stdout != nil {
		return stdout, nil, err
	}
	return stdout, stderr, err
}
