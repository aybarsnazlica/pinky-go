package main

import (
	"os"

	pinky "github.com/aybarsnazlica/pinky-go"
)

func main() {
	os.Exit(pinky.RunCLI(os.Args[1:], pinky.CLIIO{}))
}
