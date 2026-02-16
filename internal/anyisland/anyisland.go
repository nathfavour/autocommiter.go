package anyisland

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"time"
)

// Manifest info
const (
	ToolName    = "autocommiter"
	ToolRepo    = "github.com/nathfavour/autocommiter.go"
	ToolVersion = "1.0.0"
)

// Register auto-registers the tool with the local Anyisland daemon via UDP.
func Register() {
	packet := map[string]string{
		"op":      "REGISTER",
		"name":    ToolName,
		"source":  ToolRepo,
		"version": ToolVersion,
		"type":    "binary",
	}
	data, err := json.Marshal(packet)
	if err != nil {
		return
	}

	conn, err := net.DialTimeout("udp", "localhost:1995", 2*time.Second)
	if err != nil {
		return
	}
	defer conn.Close()

	_, _ = conn.Write(data)
}

// ManagedStatus represents the response from Anyisland Pulse handshake.
type ManagedStatus struct {
	Status           string `json:"status"`
	ToolID           string `json:"tool_id,omitempty"`
	Version          string `json:"version,omitempty"`
	AnyislandVersion string `json:"anyisland_version,omitempty"`
}

// CheckManaged checks if the tool is currently managed by Anyisland Pulse.
func CheckManaged() (*ManagedStatus, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	sockPath := filepath.Join(home, ".anyisland", "anyisland.sock")
	if _, err := os.Stat(sockPath); os.IsNotExist(err) {
		return &ManagedStatus{Status: "UNMANAGED"}, nil
	}

	conn, err := net.DialTimeout("unix", sockPath, 2*time.Second)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Send handshake
	handshake := map[string]string{"op": "HANDSHAKE"}
	if err := json.NewEncoder(conn).Encode(handshake); err != nil {
		return nil, err
	}

	// Read response
	var status ManagedStatus
	if err := json.NewDecoder(conn).Decode(&status); err != nil {
		return nil, err
	}

	return &status, nil
}

// IsManaged returns true if Anyisland is managing this tool.
func IsManaged() bool {
	status, err := CheckManaged()
	if err != nil {
		return false
	}
	return status.Status == "MANAGED"
}
