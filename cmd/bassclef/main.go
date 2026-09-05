package main

import (
	"fmt"
	"log"
	"os"

	"github.com/sidx04/bassclef/internal/audio"
)

func main() {
	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "info":
		runInfo(os.Args[2])
	default:
		fmt.Printf("unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("usage:")
	fmt.Println("  bassclef info <file.wav>")
}

func runInfo(path string) {
	buffer, err := audio.LoadWAV(path)
	if err != nil {
		log.Fatalf("failed to load audio: %v", err)
	}

	duration := float64(len(buffer.Samples)) /
		float64(buffer.SampleRate)

	fmt.Printf("Sample rate: %d Hz\n", buffer.SampleRate)
	fmt.Printf("Samples: %d\n", len(buffer.Samples))
	fmt.Printf("Duration: %.2f seconds\n", duration)
}
