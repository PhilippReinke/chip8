package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"unsafe"

	"github.com/PhilippReinke/chip8/pkg/display"
	"github.com/PhilippReinke/chip8/pkg/emulator"
)

const (
	instructionsPerSecond = 700
	timerHz               = 60
)

func main() {
	romPath := flag.String("rom", "", "Path to ROM")
	flag.Parse()

	chip8 := emulator.New()

	// load rom
	hexFileContent, err := os.ReadFile(*romPath)
	if err != nil {
		fmt.Printf("Failed to read ROM from %q.\n", *romPath)
		os.Exit(1)
	}
	hexBytes := parseHexFile(hexFileContent)
	chip8.LoadROM(hexBytes)

	// run app
	app := display.New(10)
	go func() {
		for range time.Tick(time.Second * 1 / instructionsPerSecond) {
			chip8.EmulateCycle()

			if chip8.DrawFlag {
				app.UpdateScreen(*(*[display.Height][display.Width]byte)(unsafe.Pointer(&chip8.Display)))
				chip8.DrawFlag = false
			}
		}
	}()
	if err := app.Run(); err != nil {
		fmt.Printf("Failed to read ROM from %q.\n", *romPath)
		os.Exit(1)
	}
}

func parseHexFile(in []byte) []byte {
	words := strings.Fields(string(in))

	var data []byte
	for _, word := range words {
		if len(word) != 4 {
			log.Fatalf("unexpected word length: %s", word)
		}
		// Decode each 4-character word (2 bytes)
		bytes, err := hex.DecodeString(word)
		if err != nil {
			log.Fatalf("failed to decode word %s: %v", word, err)
		}
		data = append(data, bytes...)
	}
	return data
}
