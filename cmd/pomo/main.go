/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"log"
	"net/http"
	_ "net/http"
	_ "net/http/pprof"

	"github.com/yardbirdsax/go-all-in/cmd/pomo/cmd"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	cmd.Execute()
}
