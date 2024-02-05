package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"editor/client/editor"
	"editor/commons"
	"editor/crdt"

	"github.com/Pallinder/go-randomdata"
	"github.com/sirupsen/logrus"
)

var (
	doc = crdt.New()

	logger = logrus.New()

	e = editor.NewEditor(editor.EditorConfig{})

	fileName string

	flags Flags
)

func main() {
	flags = parseFlags()

	s := bufio.NewScanner(os.Stdin)

	name := randomdata.SillyName()

	if flags.Login {
		fmt.Print("Enter your name: ")
		s.Scan()
		name = s.Text()
	}

	conn, _, err := createConn(flags)
	if err != nil {
		fmt.Printf("Connection error, exiting: %s\n", err)
		return
	}
	defer conn.Close()

	msg := commons.Message{Username: name, Text: "has joined the session.", Type: commons.JoinMessage}
	_ = conn.WriteJSON(msg)

	logFile, debugLogFile, err := setupLogger(logger)
	if err != nil {
		fmt.Printf("Failed to setup logger, exiting: %s\n", err)
		return
	}
	defer closeLogFiles(logFile, debugLogFile)

	if flags.File != "" {
		if doc, err = crdt.Load(flags.File); err != nil {
			fmt.Printf("failed to load document: %s\n", err)
			return
		}
	}

	uiConfig := UIConfig{
		EditorConfig: editor.EditorConfig{
			ScrollEnabled: flags.Scroll,
		},
	}

	err = initUI(conn, uiConfig)
	if err != nil {
		if strings.HasPrefix(err.Error(), "editor") {
			fmt.Println("exiting session.")
			return
		}

		fmt.Printf("TUI error, exiting: %s\n", err)
		return
	}
}
