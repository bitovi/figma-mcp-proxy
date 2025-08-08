package util

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"time"
)

// OpenFigmaDesign opens a Figma design document using the figma:// URL scheme
// On macOS: uses "open figma://design/{fileKey}/{fileName}"
// On Windows: uses "Start-Process figma://design/{fileKey}/{fileName}"
func OpenFigmaDesign(fileKey, fileName string) error {
	figmaURL := fmt.Sprintf("figma://design/%s/%s", fileKey, fileName)

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin": // macOS
		cmd = exec.Command("open", figmaURL)
	case "windows":
		cmd = exec.Command("powershell", "-Command", fmt.Sprintf("Start-Process '%s'", figmaURL))
	case "linux":
		cmd = exec.Command("xdg-open", figmaURL)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to open Figma design '%s/%s': %v", fileKey, fileName, err)
	}

	// sleep for 2 seconds to allow Figma to launch before any subsequent commands
	time.Sleep(2 * time.Second)

	log.Printf("[OSUTIL] Successfully opened Figma design '%s/%s'", fileKey, fileName)
	return nil
}

// OpenFigma opens the Figma application
// On macOS: uses "open figma://"
// On Windows: uses "Start-Process figma://"
func OpenFigma() error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin": // macOS
		cmd = exec.Command("open", "figma://")
	case "windows":
		cmd = exec.Command("powershell", "-Command", "Start-Process 'figma://'")
	case "linux":
		cmd = exec.Command("xdg-open", "figma://")
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to open Figma application: %v", err)
	}

	// sleep for 2 seconds to allow Figma to launch before any subsequent commands
	time.Sleep(2 * time.Second)

	log.Printf("[OSUTIL] Successfully opened Figma application")
	return nil
}
