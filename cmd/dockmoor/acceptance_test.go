package main

import (
	"bytes"
	"github.com/jessevdk/go-flags"
	"github.com/mattn/go-shellwords"
	"github.com/stretchr/testify/assert"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func dockerfile(content string) (fileName string) {
	contentBytes := []byte(content)
	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		log.Fatal(err)
	}

	if _, err := tmpfile.Write(contentBytes); err != nil {
		log.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
	}

	return tmpfile.Name()
}

func MainOptionsTestNew(commandAdders ...func(mainOptions *mainOptions) (*flags.Command, error)) *mainOptions {
	mainOptions := MainOptionsNew()

	mainOptions.SetStdout(bytes.NewBuffer(nil))

	for _, adder := range commandAdders {
		adder(mainOptions)
	}

	return mainOptions
}

func TestContainAnyMatches(t *testing.T) {
	df1 := dockerfile(`FROM nginx`)
	defer os.Remove(df1)

	os.Args = []string{"exe", "contains", "--any", df1}
	mainOptions := MainOptionsTestNew(addContainsCommand)

	exitCode := doMain(mainOptions)

	assert.Equal(t, EXIT_SUCCESS, exitCode)
}

func TestContainsAnyNoMatch(t *testing.T) {
	df1 := dockerfile(`invalid`)
	defer os.Remove(df1)

	os.Args = []string{"exe", "contains", "--any", df1}
	mainOptions := MainOptionsTestNew(addContainsCommand)
	exitCode := doMain(mainOptions)

	assert.NotEqual(t, EXIT_SUCCESS, exitCode)
}

func TestContainLatestMatches(t *testing.T) {
	df1 := dockerfile(`FROM nginx`)
	defer os.Remove(df1)

	os.Args = []string{"exe", "contains", "--latest", df1}
	mainOptions := MainOptionsTestNew(addContainsCommand)

	exitCode := doMain(mainOptions)

	assert.Equal(t, EXIT_SUCCESS, exitCode)
}

func TestContainsLatestNoMatch(t *testing.T) {
	df1 := dockerfile(`FROM nginx:1`)
	defer os.Remove(df1)

	os.Args = []string{"exe", "contains", "--latest", df1}
	mainOptions := MainOptionsTestNew(addContainsCommand)
	exitCode := doMain(mainOptions)

	assert.Equal(t, EXIT_NOT_FOUND, exitCode)
}

func TestContainUnpinnedMatches(t *testing.T) {
	df1 := dockerfile(`FROM nginx:1`)
	defer os.Remove(df1)

	os.Args = []string{"exe", "contains", "--unpinned", df1}
	mainOptions := MainOptionsTestNew(addContainsCommand)

	exitCode := doMain(mainOptions)

	assert.Equal(t, EXIT_SUCCESS, exitCode)
}

func TestContainsUnpinnedNoMatch(t *testing.T) {
	df1 := dockerfile(`FROM nginx:1.2@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240`)
	defer os.Remove(df1)

	os.Args = []string{"exe", "contains", "--latest", df1}
	mainOptions := MainOptionsTestNew(addContainsCommand)
	exitCode := doMain(mainOptions)

	assert.Equal(t, EXIT_NOT_FOUND, exitCode)
}

func TestContainsInvalidOptions(t *testing.T) {
	df1 := dockerfile(`FROM nginx`)
	defer os.Remove(df1)

	os.Args = []string{"exe", "contains", "--any", "--latest", df1}

	mainOptions := MainOptionsTestNew(addContainsCommand)
	exitCode := doMain(mainOptions)

	assert.Equal(t, EXIT_INVALID_PARAMS, exitCode)
}

func TestMainVersion(t *testing.T) {

	os.Args = []string{"exe", "--version"}

	mainOptions := MainOptionsTestNew(addContainsCommand)
	exitCode := doMain(mainOptions)

	assert.Equal(t, EXIT_SUCCESS, exitCode)
}

func TestMainManpage(t *testing.T) {

	os.Args = []string{"exe", "--manpage"}
	mainOptions := MainOptionsTestNew(addContainsCommand)
	exitCode := doMain(mainOptions)

	assert.Equal(t, EXIT_SUCCESS, exitCode)
}

func TestMainMarkdown(t *testing.T) {

	os.Args = []string{"exe", "--markdown"}

	mainOptions := MainOptionsTestNew()
	exitCode := doMain(mainOptions)

	assert.Equal(t, EXIT_SUCCESS, exitCode)
}

func TestMainLoglevelNone(t *testing.T) {

	os.Args = []string{"exe", "--log-level=NONE", "contains", "--any", "/notExistingFile"}
	buffer := bytes.NewBuffer(nil)
	mainOptions := MainOptionsTestNew(addContainsCommand)
	exitCode := doMain(mainOptions)

	assert.Empty(t, buffer.String())
	assert.NotEqual(t, EXIT_SUCCESS, exitCode)
}

func TestExitCodeIsZeroForDockerfile(t *testing.T) {
	dir, _ := ioutil.TempDir("", "dockmoor")
	defer os.RemoveAll(dir)

	tmpfn := filepath.Join(dir, "Dockerfile")
	dockerfile :=
		`FROM nginx`

	if err := ioutil.WriteFile(tmpfn, []byte(dockerfile), 0666); err != nil {
		log.Fatal(err)
	}

	stdout, code := shell(t, `dockmoor contains --any {{.Dockerfile}}`, struct {
		Dockerfile string
	}{tmpfn})

	assert.Empty(t, stdout)
	assert.Equal(t, EXIT_SUCCESS, code, "Exits with code 0")
}

func TestExitCodeIsZeroForContainsLatestAndDockerfile(t *testing.T) {
	dir, _ := ioutil.TempDir("", "dockmoor")
	defer os.RemoveAll(dir)

	tmpfn := filepath.Join(dir, "Dockerfile")
	dockerfile :=
		`FROM nginx`

	if err := ioutil.WriteFile(tmpfn, []byte(dockerfile), 0666); err != nil {
		log.Fatal(err)
	}

	stdout, code := shell(t, `dockmoor contains --latest {{.Dockerfile}}`, struct {
		Dockerfile string
	}{tmpfn})

	assert.Empty(t, stdout)
	assert.Equal(t, EXIT_SUCCESS, code, "Exits with code 0")
}

func TestExitCodeIsZeroForListLatestAndDockerfile(t *testing.T) {
	dir, _ := ioutil.TempDir("", "dockmoor")
	defer os.RemoveAll(dir)

	tmpfn := filepath.Join(dir, "Dockerfile")
	dockerfile :=
		`FROM nginx`

	if err := ioutil.WriteFile(tmpfn, []byte(dockerfile), 0666); err != nil {
		log.Fatal(err)
	}

	stdout, code := shell(t, `dockmoor list --latest {{.Dockerfile}}`, struct {
		Dockerfile string
	}{tmpfn})

	assert.Equal(t, "nginx\n", stdout)
	assert.Equal(t, EXIT_SUCCESS, code, "Exits with code 0")
}

func TestExitCodeIsNotZeroForInvalidDockerfile(t *testing.T) {
	dir, _ := ioutil.TempDir("", "dockmoor")
	defer os.RemoveAll(dir)

	tmpfn := filepath.Join(dir, "Dockerfile")
	dockerfile :=
		`Not from nginx`

	if err := ioutil.WriteFile(tmpfn, []byte(dockerfile), 0666); err != nil {
		log.Fatal(err)
	}

	stdout, code := shell(t, `dockmoor contains --any {{.Dockerfile}}`, struct {
		Dockerfile string
	}{tmpfn})

	assert.Empty(t, stdout)
	assert.Equal(t, EXIT_INVALID_FORMAT, code, "Exits with code 4")
}

func shell(t *testing.T, argsLine string, values interface{}) (stdout string, exitCode ExitCode) {
	tpl, _ := template.New("name").Parse(argsLine)
	shellBuf := bytes.NewBuffer(nil)
	tpl.Execute(shellBuf, values)

	argsLine = shellBuf.String()
	args, _ := shellwords.Parse(argsLine)
	exitCodeSet := false
	oldOsExit := osExitInternal
	osExitInternal = func(code int) {
		exitCode = ExitCode(code)
		exitCodeSet = true
	}

	oldStdout := osStdout
	stdoutBuf := bytes.NewBuffer(nil)
	osStdout = stdoutBuf
	oldStdin := os.Stdin
	osStdin = ioutil.NopCloser(strings.NewReader(""))
	oldArgs := os.Args
	os.Args = args

	defer func() {
		osStdout = oldStdout
		osStdin = oldStdin
		osExitInternal = oldOsExit
		os.Args = oldArgs
	}()

	main()

	buffer := bytes.NewBuffer(nil)
	buffer.ReadFrom(stdoutBuf)
	stdout = buffer.String()

	assert.True(t, exitCodeSet, "Expected exitCode to be set (no call to osExit)")
	return
}
