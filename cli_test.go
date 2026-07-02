package pinky

import "testing"

func TestRunCLIPrintsUsageWhenNoFileIsProvided(t *testing.T) {
	stdout := ""
	stderr := ""

	exitCode := RunCLI([]string{}, CLIIO{
		ReadTextFile: func(path string) (string, error) { return "", nil },
		WriteStdout:  func(text string) { stdout += text },
		WriteStderr:  func(text string) { stderr += text },
	})

	if exitCode != 1 {
		t.Fatalf("exit code = %d", exitCode)
	}
	if stdout != "" {
		t.Fatalf("stdout = %q", stdout)
	}
	if stderr != "Usage: pinky-go <file.pinky> [--verbose]\n" {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestRunCLIRunsAFileAndWritesProgramOutput(t *testing.T) {
	stdout := ""
	stderr := ""

	exitCode := RunCLI([]string{"demo.pinky"}, CLIIO{
		ReadTextFile: func(path string) (string, error) { return "println 'hello'\n", nil },
		WriteStdout:  func(text string) { stdout += text },
		WriteStderr:  func(text string) { stderr += text },
	})

	if exitCode != 0 {
		t.Fatalf("exit code = %d", exitCode)
	}
	if stdout != "hello\n" {
		t.Fatalf("stdout = %q", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q", stderr)
	}
}
