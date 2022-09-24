package main

import (
	"log"
	"os"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/yardbirdsax/go-all-in/pomo"
)

func main() {
	m := pomo.NewPomo(60 * time.Second)

	if err := tea.NewProgram(m).Start(); err != nil {
		log.Printf("error starting program: %v", err)
		os.Exit(1)
	}
}