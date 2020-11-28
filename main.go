package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/thesephist/vanta/vanta"
)

const Version = "0.1"

func main() {
	flag.Usage = func() {
		fmt.Println("Vanta is a fast Klisp interpreter.")
		flag.PrintDefaults()
	}
	interactive := flag.Bool("i", false, "Start a repl after executing files")

	// Collect and write flags
	flag.Parse()
	args := flag.Args()

	// Create execution environment
	env := vanta.New()

	// If given a file, exec it
	if len(args) > 0 {
		for _, runPath := range args {
			runFile, err := ioutil.ReadFile(runPath)
			if err != nil {
				fmt.Printf("Could not open %s: %s\n", runPath, err.Error())
			} else {
				env.Eval(vanta.Read(string(runFile)))
			}
		}
	}

	// If we should open a repl, start a read-eval-print loop
	if *interactive || len(args) == 0 {
		// REPL
		fmt.Printf("Klisp interpreter v%s-vanta.\n", Version)
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
