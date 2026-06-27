package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("uso: ocr-test <imagen>")
	}

	imagePath := os.Args[1]
	outBase := imagePath + "-ocr-out"

	cmd := exec.Command("tesseract", imagePath, outBase, "-l", "spa+eng")
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Fatalf("tesseract: %v: %s", err, string(out))
	}
	defer os.Remove(outBase + ".txt")

	data, err := os.ReadFile(outBase + ".txt")
	if err != nil {
		log.Fatalf("leer salida: %v", err)
	}

	fmt.Println(strings.TrimSpace(string(data)))
}
