package main

import (
	"log"
	"os"
	"os/exec"
	"path"
	"time"
)

func main() {
	root, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	build := path.Join(root, "build")
	if err := os.MkdirAll(build, 0755); err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile(path.Join(build, ".gitignore"), []byte("*\n"), 0622)
	if err != nil {
		log.Fatal(err)
	}

	switch os.Args[1] {
	case "commit":
		cmd := exec.Command(
			"git", "describe", "HEAD", "--tags", "--always",
		)
		output, err := cmd.Output()
		if err != nil || len(output) < 1 {
			output = []byte("current")
		} else {
			output = output[:len(output)-1]
		}

		if err := os.WriteFile(path.Join(build, "version.txt"), output, 0622); err != nil {
			log.Fatal(err)
		}
	case "date":
		output := []byte(time.Now().Format(time.RFC3339))
		if err := os.WriteFile(path.Join(build, "datetime.txt"), output, 0622); err != nil {
			log.Fatal(err)
		}
	case "compiler":
		cmd := exec.Command(
			"go", "version",
		)
		output, err := cmd.Output()
		if err != nil || len(output) < 1 {
			output = []byte("current")
		} else {
			output = output[:len(output)-1]
		}

		if err := os.WriteFile(path.Join(build, "compiler.txt"), output, 0622); err != nil {
			log.Fatal(err)
		}
	}
}
