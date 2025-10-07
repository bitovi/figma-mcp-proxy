package util

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func escapeColonsForFigma(nodeId string) string {
	log.Printf("[UTIL] Escaping colons in nodeId: %s", nodeId)
	escaped := strings.ReplaceAll(nodeId, ":", "-")
	log.Printf("[UTIL] Escaped nodeId: %s", escaped)
	return escaped
}

// OpenFigmaDesign opens a Figma design document using the figma:// URL scheme
// On macOS: uses "open figma://design/{fileKey}/{fileName}"
// On Windows: uses "Start-Process figma://design/{fileKey}/{fileName}"
func OpenFigmaDesign(fileKey, fileName, nodeId string) error {
	log.Printf("[UTIL] OpenFigmaDesign called with fileKey: %s, fileName: %s, nodeId: %s", fileKey, fileName, nodeId)

	escapedNodeId := escapeColonsForFigma(nodeId)
	figmaURL := fmt.Sprintf("figma://design/%s/%s?node-id=%s", fileKey, fileName, escapedNodeId)
	log.Printf("[UTIL] Generated Figma URL: %s", figmaURL)

	var cmd *exec.Cmd
	osType := runtime.GOOS
	log.Printf("[UTIL] Detected operating system: %s", osType)

	switch osType {
	case "darwin": // macOS
		log.Printf("[UTIL] Using macOS 'open' command")
		cmd = exec.Command("open", figmaURL)
	case "windows":
		log.Printf("[UTIL] Using Windows PowerShell Start-Process command")
		cmd = exec.Command("powershell", "-Command", fmt.Sprintf("Start-Process '%s'", figmaURL))
	case "linux":
		log.Printf("[UTIL] Using Linux 'xdg-open' command")
		cmd = exec.Command("xdg-open", figmaURL)
	default:
		log.Printf("[UTIL] ERROR: Unsupported operating system: %s", osType)
		return fmt.Errorf("unsupported operating system: %s", osType)
	}

	log.Printf("[UTIL] Executing command: %s %v", cmd.Path, cmd.Args)
	err := cmd.Run()
	if err != nil {
		log.Printf("[UTIL] ERROR: Failed to execute command: %v", err)
		return fmt.Errorf("failed to open Figma design '%s/%s': %v", fileKey, fileName, err)
	}
	log.Printf("[UTIL] Command executed successfully")

	// sleep for 2 seconds to allow Figma to launch before any subsequent commands
	log.Printf("[UTIL] Sleeping for 2 seconds to allow Figma to launch")
	time.Sleep(2 * time.Second)
	log.Printf("[UTIL] Sleep completed, OpenFigmaDesign finished successfully")

	return nil
}

// OpenFigma opens the Figma application
// On macOS: uses "open figma://"
// On Windows: uses "Start-Process figma://"
func OpenFigma() error {
	log.Printf("[UTIL] OpenFigma called to launch Figma application")

	var cmd *exec.Cmd
	osType := runtime.GOOS
	log.Printf("[UTIL] Detected operating system: %s", osType)

	switch osType {
	case "darwin": // macOS
		log.Printf("[UTIL] Using macOS 'open' command for figma://")
		cmd = exec.Command("open", "figma://")
	case "windows":
		log.Printf("[UTIL] Using Windows PowerShell Start-Process command for figma://")
		cmd = exec.Command("powershell", "-Command", "Start-Process 'figma://'")
	case "linux":
		log.Printf("[UTIL] Using Linux 'xdg-open' command for figma://")
		cmd = exec.Command("xdg-open", "figma://")
	default:
		log.Printf("[UTIL] ERROR: Unsupported operating system: %s", osType)
		return fmt.Errorf("unsupported operating system: %s", osType)
	}

	log.Printf("[UTIL] Executing command: %s %v", cmd.Path, cmd.Args)
	err := cmd.Run()
	if err != nil {
		log.Printf("[UTIL] ERROR: Failed to execute command: %v", err)
		return fmt.Errorf("failed to open Figma application: %v", err)
	}
	log.Printf("[UTIL] Command executed successfully")

	// sleep for 2 seconds to allow Figma to launch before any subsequent commands
	log.Printf("[UTIL] Sleeping for 2 seconds to allow Figma to launch")
	time.Sleep(2 * time.Second)
	log.Printf("[UTIL] Sleep completed, OpenFigma finished successfully")

	return nil
}
