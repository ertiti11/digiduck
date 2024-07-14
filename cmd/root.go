package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "digiducky",
	Short: "Una herramienta para codificar archivos en formato ducky y generar sketches para Digispark",
}

// Execute ejecuta el comando raíz.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Aquí puedes definir banderas y configuraciones para el comando raíz.
}
