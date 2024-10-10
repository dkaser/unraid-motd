package utils

import (
	"encoding/json"
	"fmt"
)

const (
	tebibyte float64 = 1099511627776
	gibibyte float64 = 1073741824
	mebibyte float64 = 1048576
	kebibyte float64 = 1024
)

// FormatBytes format bytes to TiB, GiB, MiB, KiB depending on the size
func FormatBytes(sizeBytes float64) string {
	if sizeBytes > tebibyte {
		return fmt.Sprintf("%.2f TB", sizeBytes/tebibyte)
	}
	if sizeBytes > gibibyte {
		return fmt.Sprintf("%.2f GB", sizeBytes/gibibyte)
	}
	if sizeBytes > mebibyte {
		return fmt.Sprintf("%.2f MB", sizeBytes/mebibyte)
	}
	if sizeBytes > kebibyte {
		return fmt.Sprintf("%.2f KB", sizeBytes/kebibyte)
	}

	return fmt.Sprintf("%.2f B", sizeBytes)
}

// PrettyPrint a struct, used for debugging
func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", " ")

	return string(s)
}
