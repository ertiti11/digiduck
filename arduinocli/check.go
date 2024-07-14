package arduinocli

import (
	"digiduck/utils"
	"os"
	"os/exec"
	"path/filepath"
)

func isArduinoCLIInstalledOnWindows() bool {
	path := filepath.Join(`C:\Arduino CLI`, "arduino-cli.exe")
	_, err := os.Stat(path)
	return err == nil
}

// IsArduinoCLIInstalled verifica si arduino-cli está instalado en el sistema.
func IsArduinoCLIInstalled() bool {
	if utils.GetOSName() == "windows" {
		return isArduinoCLIInstalledOnWindows()
	}
	_, err := exec.LookPath("arduino-cli")
	return err == nil
}
