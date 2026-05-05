package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func main() {
	inputFile := flag.String("input", "", "Archivo JSON de entrada (formato legacy)")
	outputFile := flag.String("output", "tiendas_import.json", "Archivo JSON de salida")
	galeria := flag.String("gal", "torre-flamenco", "Galería: placa-comercial | torre-flamenco")
	flag.Parse()

	if *galeria != "placa-comercial" && *galeria != "torre-flamenco" {
		fmt.Fprintf(os.Stderr, "Error: gal debe ser 'placa-comercial' o 'torre-flamenco'\n")
		os.Exit(1)
	}

	if *inputFile == "" {
		fmt.Fprintln(os.Stderr, "Uso: scraper -input <archivo.json> -gal placa-comercial|torre-flamenco")
		os.Exit(1)
	}

	data, err := os.ReadFile(*inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error leyendo archivo: %v\n", err)
		os.Exit(1)
	}

	tiendas, err := transformJSON(data, *galeria)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error transformando JSON: %v\n", err)
		os.Exit(1)
	}

	out, err := json.MarshalIndent(tiendas, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error serializando JSON: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(*outputFile, out, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error escribiendo output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%d tiendas escritas en %s\n", len(tiendas), *outputFile)
}
