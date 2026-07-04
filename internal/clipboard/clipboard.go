package clipboard

import (
	"os/exec"
	"runtime"
	"strings"
)

// Copy writes text to the system clipboard.
// Returns an error if no clipboard tool is available.
func Copy(text string) error {
	switch runtime.GOOS {
	case "windows":
		return run("clip", nil, text)
	case "darwin":
		return run("pbcopy", nil, text)
	default:
		if err := run("xclip", []string{"-selection", "clipboard"}, text); err == nil {
			return nil
		}
		return run("xsel", []string{"--clipboard", "--input"}, text)
	}
}

func run(name string, args []string, input string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = strings.NewReader(input)
	return cmd.Run()
}
