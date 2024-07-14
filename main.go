package main

import (
	"digiduck/arduinocli"
	"digiduck/cmd"
	"fmt"
	"log"
)

func main() {
	if arduinocli.IsArduinoCLIInstalled() {
		fmt.Println("arduino-cli are installed")
		cmd.Execute()
	} else {
		fmt.Println("arduino-cli are not installed")
		// Puedes implementar una lógica para instalarlo automáticamente aquí
		log.Println("trying to install arduino-cli...")
		if err := arduinocli.InstallArduinoCLI(); err != nil {
			log.Fatalf("Error installing arduino-cli: %v", err)
		}
		fmt.Println("arduino-cli has been installed correctly")
	}
}
