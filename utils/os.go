package utils

import (
	"runtime"
)

// GetOSName devuelve el nombre del sistema operativo actual.
func GetOSName() string {
	return runtime.GOOS
}
