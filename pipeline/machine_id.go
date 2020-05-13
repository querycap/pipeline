package pipeline

import (
	"path/filepath"
	"strings"
)

type MachineIdentifier interface {
	MachineID() (string, error)
}

func FilenameWithMachineID(machineID string, filename string) string {
	if machineID != "" {
		return machineID + "," + filename
	}
	return filename
}

func GetMachineIDFromFilename(filename string) string {
	i := strings.Index(filepath.Base(filename), ",")
	if i > 0 {
		return filename[0:i]
	}
	return ""
}
