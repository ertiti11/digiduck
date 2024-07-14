package arduinocli

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// IsArduinoCLIInstalled verifica si arduino-cli está instalado en el sistema.
const (
	arduinoCLIDownloadURL = "https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_Windows_64bit.zip"
)


// InstallArduinoCLI intenta instalar arduino-cli según el sistema operativo.
func InstallArduinoCLI() error {
	switch runtime.GOOS {
	case "linux":
		return InstallOnLinux()
	case "darwin":
		return InstallOnDarwin()
	case "windows":
		return InstallOnWindows()
	default:
		return fmt.Errorf("sistema operativo no soportado: %s", runtime.GOOS)
	}
}

// InstallOnLinux instala arduino-cli en sistemas Linux.
func InstallOnLinux() error {
	// Implementa la lógica para instalar arduino-cli en Linux aquí.
	// Por ejemplo, puedes usar un gestor de paquetes como apt-get o descargar directamente desde el sitio web oficial.
	log.Println("Instalando arduino-cli en Linux...")
	return exec.Command("sudo", "apt-get", "install", "arduino-cli").Run()
}

// InstallOnDarwin instala arduino-cli en macOS.
func InstallOnDarwin() error {
	// Implementa la lógica para instalar arduino-cli en macOS aquí.
	// Puedes usar Homebrew u otros métodos de instalación.
	log.Println("Instalando arduino-cli en macOS...")
	return exec.Command("brew", "install", "arduino-cli").Run()
}

func InstallOnWindows() error {
	log.Println("Descargando arduino-cli para Windows...")
	zipFile := filepath.Join(os.TempDir(), "arduino-cli_latest_Windows_64bit.zip")

	// Descargar el archivo ZIP desde el sitio oficial
	err := downloadFile(arduinoCLIDownloadURL, zipFile)
	if err != nil {
		return fmt.Errorf("error descargando arduino-cli: %v", err)
	}
	defer os.Remove(zipFile)

	log.Println("Descomprimiendo arduino-cli...")

	// Crear un directorio temporal para extraer el contenido del ZIP
	extractDir := filepath.Join(os.TempDir(), "arduino-cli_extracted")
	if err := os.MkdirAll(extractDir, os.ModePerm); err != nil {
		return fmt.Errorf("error creando directorio temporal: %v", err)
	}
	defer os.RemoveAll(extractDir)

	// Extraer el archivo ZIP
	if err := extractZip(zipFile, extractDir); err != nil {
		return fmt.Errorf("error extrayendo arduino-cli: %v", err)
	}

	// Definir la ubicación de destino en C:\Program Files\Arduino CLI\
	destDir := "C:\\Arduino CLI"
	destPath := filepath.Join(destDir, "arduino-cli.exe")
	srcPath := filepath.Join(extractDir, "arduino-cli.exe")

	// Verificar si el directorio de destino existe, si no, crearlo
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
			return fmt.Errorf("error creando directorio %s: %v", destDir, err)
		}
	}

	// Mover arduino-cli.exe a la ubicación de destino
	if err := os.Rename(srcPath, destPath); err != nil {
		return fmt.Errorf("error moviendo arduino-cli a %s: %v", destPath, err)
	}

	log.Printf("arduino-cli instalado correctamente en %s\n", destPath)
	return nil
}

// downloadFile descarga un archivo desde una URL y lo guarda en el sistema de archivos.
func downloadFile(url, filePath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// extractZip extrae un archivo ZIP en una ubicación específica.
func extractZip(zipFile, dest string) error {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, os.ModePerm)
		} else {
			if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
				return err
			}
			outFile, err := os.Create(path)
			if err != nil {
				return err
			}
			defer outFile.Close()

			_, err = io.Copy(outFile, rc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}