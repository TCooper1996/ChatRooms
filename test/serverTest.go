package main

import (
	"log"
	"os"
	"os/exec"
)

func main() {

	cmd := exec.Command("./go_build")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); nil != err {
		log.Fatalf("error starting: %s", err.Error())
	}

	cmd.Wait()

}
