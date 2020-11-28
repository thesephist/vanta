package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/thesephist/vanta/vanta"
)

func main() {
	env := vanta.New()

	// if given a file, exec it
	for _, runPath := range os.Args[1:] {
		runFile, err := ioutil.ReadFile(runPath)
		if err != nil {
			fmt.Printf("Could not open %s: %s\n", runPath, err.Error())
		} else {
			env.Eval(vanta.Read(string(runFile)))
		}
	}

	stdin, _ := os.Stdin.Stat()
	if (stdin.Mode() & os.ModeCharDevice) != 0 {
		// REPL
		fmt.Println("Klisp interpreter v0.1-vanta.")
		reader := bufio.NewReader(os.Stdin)

		for {
			fmt.Printf("> ")
			text, err := reader.ReadString('\n')
			if err == io.EOF {
				break
			} else if err != nil {
				log.Fatal("Unexpected end of input", err)
			}

			val := env.Eval(vanta.Read(text))
			fmt.Println(vanta.Print(val))
		}
	}
}
