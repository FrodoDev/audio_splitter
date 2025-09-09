package main

import (
	"fmt"
	"os"
)

var homeDir string

func init() {
	homeDir, _ = os.UserHomeDir()
	fmt.Println("homeDir:", homeDir)
}

func main() {
	xlsxPath := "./templet.xlsx"
	// SplitAudio(xlsxPath)
	AddLyrics(xlsxPath)
}
