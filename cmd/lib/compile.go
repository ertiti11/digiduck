package compiler

import(
	"log"
	"os"
	"os/exec"
	"github.com/fatih/color"
)

func compile() {
	color.Yellow("compilando...\n")
	compile := exec.Command("./lib/arduino.exe", "compile", "-b", "digistump:avr:digispark-tiny")
	compile_rr := compile.Run()
	if compile_rr != nil {
		color.Red("no se pudo compilar el sketch")
		os.Exit(1)
	}
	color.Green("compilado \n")
	color.Green("Tienes 60 seg para introducir el dispositivo \n")
	// go func() {
	// 	time.Sleep(2 * time.Second)
	// 	color.Green("Introduce el digispark en un puerto usb (tiempo restante 60 seg....")

	// }()

	upload := exec.Command("./lib/arduino.exe", "upload", "-b", "digistump:avr:digispark-tiny")
	upload_err := upload.Run()
	if upload_err != nil {
		log.Fatal(upload_err)
	}
	color.Green("subido? \n")

}