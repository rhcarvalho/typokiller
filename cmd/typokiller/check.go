package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"

	"github.com/rhcarvalho/typokiller/pkg/types"
)

// Check spell checks the locations in the current project.
func (m *Main) Check() error {
	cmd, cmdStdin, cmdStdout, err := startSpellCheck()
	if err != nil {
		return fmt.Errorf("cannot start spell checker: %v", err)
	}

	stdoutEnc := json.NewEncoder(m.Stdout)
	stdinDec := json.NewDecoder(m.Stdin)
	c := make(chan error)

	// Send locations to the spell checker in a separate goroutine.
	go func() {
		c <- func() error {
			spellCheckEnc := json.NewEncoder(cmdStdin)
			var pkg types.Package
			for {
				// Decode pkg from stdin.
				if err := stdinDec.Decode(&pkg); err == io.EOF {
					break
				} else if err != nil {
					return fmt.Errorf("decode pkg from stdin: %v", err)
				}
				// Encode pkg to spell checker.
				if err := spellCheckEnc.Encode(pkg); err != nil {
					return fmt.Errorf("encode pkg to spell checker: %v", err)
				}
			}
			// Done writing.
			return cmdStdin.Close()
		}()
	}()

	spellCheckDec := json.NewDecoder(cmdStdout)
	for {
		// Decode pkg from spell checker.
		var pkg types.Package
		if err := spellCheckDec.Decode(&pkg); err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("decode pkg from spell checker: %v", err)
		}

		// Encode pkg to stdout.
		if err := stdoutEnc.Encode(pkg); err != nil {
			if err, ok := err.(*os.PathError); ok && err.Err == syscall.EPIPE {
				// Ignore broken pipe error.
				return nil
			}
			return fmt.Errorf("encode pkg to stdout: %v", err)
		}
	}

	// Wait until there are no more locations to be sent.
	if err := <-c; err != nil {
		return fmt.Errorf("send locations to spell checker: %v", err)
	}

	return cmd.Wait()
}

// startSpellCheck starts an external spell checker process and return the cmd
// to be waited for and its stdin and stdout. An error is returned if something
// goes wrong.
func startSpellCheck() (cmd *exec.Cmd, stdin io.WriteCloser, stdout io.ReadCloser, err error) {
	return startCmd("./spellcheck.py")
}

func startCmd(name string, arg ...string) (cmd *exec.Cmd, stdin io.WriteCloser, stdout io.ReadCloser, err error) {
	cmd = exec.Command(name, arg...)
	stdin, err = cmd.StdinPipe()
	if err != nil {
		return
	}
	stdout, err = cmd.StdoutPipe()
	if err != nil {
		return
	}
	if err = cmd.Start(); err != nil {
		return
	}
	return
}
