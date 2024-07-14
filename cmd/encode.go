package cmd

import (
	"bufio"
	"bytes"
	"digiduck/goduck"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// encodeCmd representa el comando encode
var encodeCmd = &cobra.Command{
	Use:   "encode [archivo] -l [layout]",
	Short: "Codifica un archivo en formato ducky",
	Long: `Codifica un archivo en formato ducky y genera un sketch para Digispark.

Ejemplo de uso:
  digiducky encode archivo.bin -l es`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		file, err := filepath.Abs(args[0])
		if err != nil {
			log.Fatalf("error obteniendo la ruta absoluta del archivo: %v", err)
		}

		fmt.Println("Antes de codificar")
		err = encode(file)
		if err != nil {
			color.Red(err.Error())
			os.Exit(2)
		}
		fmt.Println("Después de codificar")

		payload, err := getFile("war.bin")
		if err != nil {
			log.Fatalf("error al obtener el archivo war.bin: %v", err)
		}

		digiscript := generateSource(payload, 2500, 1, 5000, true)
		err = writeData(digiscript)
		if err != nil {
			log.Fatalf("error al escribir el script en sketch.ino: %v", err)
		}

		err = os.Remove("war.bin")
		if err != nil {
			fmt.Println("No se pudo eliminar war.bin:", err)
		}

		compile()
	},
}

func init() {
	rootCmd.AddCommand(encodeCmd)
	encodeCmd.Flags().StringP("layout", "l", "es", "Layout del teclado (es, eng, ...)")
}

func encode(file string) error {
	encoder := goduck.NewEncoder(file, "war.bin", "resources/es.yml")
	if err := encoder.Encode(); err != nil {
		log.Fatal(err)
	}
	return nil
}

func customEncode(digit byte) string {
	return fmt.Sprintf("0x%x", digit)
}

func generateSource(payload []byte, initDelay, loopCount int, loopDelay int, blink bool) string {
	delay := strconv.Itoa(initDelay)
	loopDel := strconv.Itoa(loopDelay)

	head := `/*
* Sketch generated by duckgo from ertiti11 with go
*
*/
#include "DigiKeyboard.h"
`
	init := `
void setup()
{
	pinMode(0, OUTPUT); // LED on Model B
	pinMode(1, OUTPUT); // LED on Model A
	DigiKeyboard.delay(` + delay + `); 
}
void loop()
{
`
	body := `
	if (i != 0) {
		DigiKeyboard.sendKeyStroke(0);
		for (int i=0; i<DUCK_LEN; i+=2)
		{
			uint8_t key = pgm_read_word_near(duckraw + i);
			uint8_t mod = pgm_read_word_near(duckraw + i+1);
			if (key == 0) {
				DigiKeyboard.delay(mod);
			} else {
				DigiKeyboard.sendKeyStroke(key, mod);
			}
		}
		i--;
		DigiKeyboard.delay(` + loopDel + `); 
	}
	else if (blink)
	{
		digitalWrite(0, HIGH);
		digitalWrite(1, HIGH);
		delay(100);
		digitalWrite(0, LOW);
		digitalWrite(1, LOW);
		delay(100);
	}
`
	tail := "}"

	l := len(payload)
	declare := "#define DUCK_LEN " + fmt.Sprintf("%d", l) + "\nconst PROGMEM uint8_t duckraw [DUCK_LEN] = {\n\t"
	for i := 0; i < l-1; i++ {
		declare += customEncode(payload[i]) + ", "
	}
	declare += customEncode(payload[l-1]) + "\n};\nint i = 1;\nbool blink=true;\n"

	return head + declare + init + body + tail
}

func writeData(digiscript string) error {
	f, err := os.Create("digiduck.ino")
	if err != nil {
		return fmt.Errorf("error creando el archivo digiduck.ino: %w", err)
	}
	defer f.Close()

	_, err = f.WriteString(digiscript)
	if err != nil {
		return fmt.Errorf("error escribiendo en digiduck.ino: %w", err)
	}
	return nil
}

func getFile(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error abriendo el archivo %s: %w", filename, err)
	}
	defer file.Close()

	stats, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo información del archivo %s: %w", filename, err)
	}

	bytes := make([]byte, stats.Size())
	_, err = bufio.NewReader(file).Read(bytes)
	if err != nil {
		return nil, fmt.Errorf("error leyendo el archivo %s: %w", filename, err)
	}

	return bytes, nil
}

func compile() {
	color.Yellow("Compilando...\n")

	compile := exec.Command("./cmd/lib/arduino.exe", "compile", "-b", "digistump:avr:digispark-tiny")
	// Captura el stderr del comando
	var stderr bytes.Buffer
	compile.Stderr = &stderr
	err := compile.Run()
	if err != nil {
		fmt.Println("Error:", err.Error())
		fmt.Println("Stderr:", stderr.String())
		// color.Red("Error al compilar el sketch")
		os.Exit(1)
	}
	color.Green("Compilado correctamente\n")
	color.Green("Tienes 60 segundos para conectar el dispositivo\n")

	time.Sleep(2 * time.Second)
	color.Green("Conecta el digispark a un puerto USB (tiempo restante 60 segundos...)\n")

	upload := exec.Command("./cmd/lib/arduino.exe", "upload", "-b", "digistump:avr:digispark-tiny")
	err = upload.Run()
	if err != nil {
		log.Fatalf("Error al subir el sketch: %v", err)
	}
	color.Green("Sketch subido correctamente\n")
}
