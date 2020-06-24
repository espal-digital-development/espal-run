package main

import (
	"bytes"
	"log"
	"os/exec"
)

func buildQTC() {
	log.Println("Building templates. Please wait..")
	out, err := exec.Command("cd", "pages", "&&", "qtc").CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	if bytes.Contains(out, []byte("error")) {
		log.Println(string(out))
	}
}
