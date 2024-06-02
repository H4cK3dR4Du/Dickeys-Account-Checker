package main

import (
	"fmt"
	"os/exec"
)

func runCommand(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to execute %s %v: %w", command, args, err)
	}
	return nil
}

func main() {
	commands := []struct {
		command string
		args    []string
	}{
		{"go", []string{"mod", "download", "github.com/fatih/color"}},
		{"go", []string{"mod", "download", "golang.org/x/sys"}},
		{"go", []string{"get", "github.com/fatih/color@v1.10.0"}},
	}

	for _, cmd := range commands {
		fmt.Printf("Running %s %v\n", cmd.command, cmd.args)
		err := runCommand(cmd.command, cmd.args...)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Printf("Successfully executed %s %v\n", cmd.command, cmd.args)
	}
}
