//go:build mage

package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/magefile/mage/mg"

	//mage:import
	"github.com/dosquad/mage"
	"github.com/dosquad/mage/helper"
	"github.com/dosquad/mage/helper/paths"
	"github.com/dosquad/mage/loga"
)

// TestLocal update, protoc, format, tidy, lint & test.
func TestLocal(ctx context.Context) {
	mg.CtxDeps(ctx, mage.Test)
	mg.CtxDeps(ctx, TestRun)
}

var Default = TestLocal

func TestRun(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	mage.Build.DebugCommand(mage.Build{}, ctx, "docker-pull-output")

	commandPaths := paths.MustCommandPaths()

	if len(commandPaths) != 1 {
		mg.Fatal(1, "unable to get command path for docker-pull-output")
	}

	cmdTemplate := helper.NewCommandTemplate(true, commandPaths[0])

	loga.PrintCommandAlways("%s", cmdTemplate.OutputArtifact)
	cmd := exec.CommandContext(ctx, cmdTemplate.OutputArtifact)

	{
		testInputFile := paths.MustGetGitTopLevel("testdata", "testoutput.txt")
		f, err := os.Open(testInputFile)
		if err != nil {
			mg.Fatalf(1, "unable to read test docker pull output (%s): %s", testInputFile, err)
		}
		defer f.Close()

		cmd.Stdin = f

		// go func() {
		// 	_, err := io.Copy(testInput, f)
		// 	if err != nil {
		// 		mg.Fatalf(1, "unable to copy from test input file to stdin: %s", err)
		// 	}
		// 	testInput.Close()
		// }()
	}

	testOutput := bytes.NewBuffer(nil)
	testError := bytes.NewBuffer(nil)

	cmd.Stdout = testOutput
	cmd.Stderr = testError

	if err := cmd.Run(); err != nil {
		mg.Fatalf(1, "unable to execute docker pull output test (%s): %s", commandPaths[0], err)
	}

	if testOutput.Len() > 0 {
		mg.Fatalf(1, "stdout output produced (only output should be on stderr): %s", testOutput.Bytes())
	}

	output := filterOutput(testError.Bytes())

	var expect []string
	{
		testExpectedFile := paths.MustGetGitTopLevel("testdata", "testoutput.txt.run")
		var f *os.File
		{
			var err error
			f, err = os.Open(testExpectedFile)
			if err != nil {
				mg.Fatalf(1, "unable to read expected docker pull output file (%s): %s", testExpectedFile, err)
			}
		}

		data, err := io.ReadAll(f)
		if err != nil {
			mg.Fatalf(1, "unable to read lines from expected docker pull output file: %s", err)
		}

		expect = filterOutput(data)
	}

	if diff := cmp.Diff(output, expect); diff != "" {
		loga.PrintWarning("docker-pull-output: -got +want:\n%s", diff)
		mg.Fatalf(1, "docker-pull-output: -got +want:\n%s", diff)
	}
}

func filterOutput(in []byte) []string {
	f := regexp.MustCompile(`.* level=`)

	lines := strings.Split(string(in), "\n")
	for idx := range lines {
		lines[idx] = f.ReplaceAllString(lines[idx], "")
	}

	return lines
}
