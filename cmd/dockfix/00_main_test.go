package main

import (
	"testing"
	"io"
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/sirupsen/logrus"
	"github.com/jessevdk/go-flags"
	"github.com/MeneDev/dockfix/dockfmt"
)

var NotADockerfile = "notDocker"

type mainOptionsTest struct {
	*MainOptions
	openerMock *ReadableOpenerMock
}

func MainOptionsTest() *mainOptionsTest {
	mainOptions := mainOptionsTest{MainOptions: &MainOptions{}}
	parser := flags.NewParser(mainOptions.MainOptions, flags.HelpFlag|flags.PassDoubleDash)
	mainOptions.parser = parser

	stdout := bytes.NewBuffer(nil)

	mainOptions.log = logrus.New()
	mainOptions.SetStdout(stdout)

	mainOptions.formatProvider = new(dockfmt.FormatProviderMock)

	mainOptions.openerMock = new(ReadableOpenerMock)
	mainOptions.openerMock.On("Open", NotADockerfile).Return(makeReadCloser("not a dockerfile"), nil)

	mainOptions.readableOpener = func(s string) (io.ReadCloser, error) {
		return mainOptions.openerMock.Open(s)
	}

	return &mainOptions
}

func (options *mainOptionsTest) FormatProvider() *dockfmt.FormatProviderMock {
	return options.formatProvider.(*dockfmt.FormatProviderMock)
}
func (options *mainOptionsTest) Stdout() *bytes.Buffer {
	return options.stdout.(*bytes.Buffer)
}

func testMain(args []string, registerOptions ...func(mainOptions *MainOptions) (*flags.Command, error)) (theCommand flags.Commander, cmdArgs []string, exitCode int, buffer *bytes.Buffer) {
	mainOptions := MainOptionsTest()

	for _, reg := range registerOptions {
		reg(mainOptions.MainOptions)
	}

	cmd, args, exitCode := doMain(mainOptions.MainOptions, args)

	return cmd, args, exitCode, mainOptions.Stdout()
}

func TestNoCommandKnownIsError(t *testing.T) {
	_, _, exitCode, stdout := testMain([]string{})
	s := stdout.String()
	assert.NotEqual(t, 0, exitCode)
	assert.Contains(t, s, "level=error")
	assert.Contains(t, s, "No Command registered")
}

func TestHelpIsNotError(t *testing.T) {
	_, _, exitCode, stdout := testMain([]string{"--help"})
	s := stdout.String()
	assert.Equal(t, 0, exitCode)
	assert.NotContains(t, s, "level=error")
	assert.Contains(t, s, "Usage")
	assert.Contains(t, s, "Application Options")
	assert.Contains(t, s, "Help Options")
}

func TestManIsNotError(t *testing.T) {
	_, _, exitCode, stdout := testMain([]string{"--manpage"})
	s := stdout.String()
	assert.Equal(t, 0, exitCode)
	assert.NotContains(t, s, "level=error")
	assert.Contains(t, s, "NAME")
	assert.Contains(t, s, "SYNOPSIS")
	assert.Contains(t, s, "OPTIONS")
}

func TestOpensStdin(t *testing.T) {

	optionsTest := MainOptionsTest()
	opener := defaultReadableOpener(optionsTest.MainOptions)

	readCloser, e := opener("-")

	assert.Nil(t, e)
	assert.Equal(t, optionsTest.stdin, readCloser)

}