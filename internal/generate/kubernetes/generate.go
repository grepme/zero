package kubernetes

import (
	"bytes"
	"github.com/commitdev/commit0/internal/config"
	"github.com/commitdev/commit0/internal/templator"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

func Generate(templator *templator.Templator, config *config.Commit0Config, wg *sync.WaitGroup) {
	templator.Kubernetes.TemplateFiles(config, false, wg)

	if config.Kubernetes.Deploy {
		_tf_init := tf_init()
		_tf_plan := tf_plan()
		execute(_tf_init)
		execute(_tf_plan)
	}

}

// Terraform init cmd
func tf_init() *exec.Cmd {

	return exec.Command("terraform", "init")
}

// Terraform plan cmd
func tf_plan() *exec.Cmd {

	return exec.Command("terraform", "plan")
}

// Executes cmd passed in
func execute(cmd *exec.Cmd) {
	dir, err1 := filepath.Abs(filepath.Dir(os.Args[0]))
	if err1 != nil {
		log.Fatal(err1)
	}

	cmd.Dir = dir + "/kubernetes/terraform/environments/staging"

	var errStdout, errStderr error
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()
	stdout := NewCapturingPassThroughWriter(os.Stdout)
	stderr := NewCapturingPassThroughWriter(os.Stderr)
	err := cmd.Start()
	if err != nil {
		log.Fatalf("cmd.Start() failed with '%s'\n", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		_, errStdout = io.Copy(stdout, stdoutIn)
		wg.Done()
	}()

	_, errStderr = io.Copy(stderr, stderrIn)
	wg.Wait()

	err = cmd.Wait()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	if errStdout != nil || errStderr != nil {
		log.Fatal("failed to capture stdout or stderr\n")
	}
}

// CapturingPassThroughWriter is a writer that remembers
// data written to it and passes it to w
type CapturingPassThroughWriter struct {
	buf bytes.Buffer
	w   io.Writer
}

// NewCapturingPassThroughWriter creates new CapturingPassThroughWriter
func NewCapturingPassThroughWriter(w io.Writer) *CapturingPassThroughWriter {
	return &CapturingPassThroughWriter{
		w: w,
	}
}

// Write writes data to the writer, returns number of bytes written and an error
func (w *CapturingPassThroughWriter) Write(d []byte) (int, error) {
	w.buf.Write(d)
	return w.w.Write(d)
}

// Bytes returns bytes written to the writer
func (w *CapturingPassThroughWriter) Bytes() []byte {
	return w.buf.Bytes()
}
