package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func launchDesktopFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var execLine string
	for line := range strings.SplitSeq(string(data), "\n") {
		if after, ok := strings.CutPrefix(line, "Exec="); ok {
			execLine = strings.TrimSpace(after)
			break
		}
	}
	if execLine == "" {
		return fmt.Errorf("no Exec line found in %s", path)
	}

	// Remove only placeholders %f %F %u %U %i %c %k
	placeholders := []string{"%f", "%F", "%u", "%U", "%i", "%c", "%k"}
	for _, ph := range placeholders {
		execLine = strings.ReplaceAll(execLine, ph, "")
	}
	execLine = strings.TrimSpace(execLine)

	// Use /bin/sh -c to handle quoted args
	cmd := exec.Command("/bin/sh", "-c", execLine)

	// Detach from parent terminal
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if os.Getenv("GREG_DRY_RUN") == "1" {
		fmt.Fprintln(os.Stdout, "DRY-RUN-EXEC:", execLine)
		return nil
	}
	return cmd.Start()
}
