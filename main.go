package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
)

const FREEZE_FILE = ".freeze"

type CommandFunc func(cmd *Command, args []string)

type Command struct {
	Run         CommandFunc
	Description string
}

var commands = map[string]*Command{
	"init": &Command{
		Run:         runInit,
		Description: "initialize freeze file",
	},
	"check": &Command{
		Run:         runCheck,
		Description: "validate files",
	},
	"update": &Command{
		Run:         runUpdate,
		Description: "update freeze file",
	},
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func usage() {
	fmt.Println("Usage: freeze [COMMAND]")
	fmt.Println("Commands:")
	for name, cmd := range commands {
		fmt.Printf("    %-10s%s\n", name, cmd.Description)
	}
	os.Exit(2)
}

func runInit(cmd *Command, args []string) {
	flags := flag.NewFlagSet("init", flag.ExitOnError)
	force := flags.Bool("force", false, "overwrite freeze file")
	flags.Parse(args)

	if !*force {
		_, err := os.Stat(FREEZE_FILE)
		if err == nil || !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Freeze file already exists, use -force to overwrite.\n")
			os.Exit(1)
		}
	}

	results := generateFreeze()
	b, err := json.MarshalIndent(results, "", "    ")
	if err != nil {
		panic(err)
	}
	output, err := os.Create(FREEZE_FILE)
	if err != nil {
		panic(err)
	}
	defer output.Close()
	output.Write(b)
	output.Write([]byte("\n"))
}

func runCheck(cmd *Command, args []string) {
	flags := flag.NewFlagSet("check", flag.ExitOnError)
	flags.Parse(args)

	verifyFreeze(getFreeze())
}

func runUpdate(cmd *Command, args []string) {
	flags := flag.NewFlagSet("update", flag.ExitOnError)
	flags.Parse(args)
}

func getFreeze() map[string]string {
	f, err := os.Open(FREEZE_FILE)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		panic(nil)
	}
	var expectedResults map[string]string
	err = json.Unmarshal(data, &expectedResults)
	if err != nil {
		panic(err)
	}
	return expectedResults
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		usage()
	}

	if cmd, ok := commands[args[0]]; ok {
		cmd.Run(cmd, args[1:])
	} else {
		fmt.Fprintf(os.Stderr, "Unknown command '%s'.\n", args[0])
		usage()
	}
}
