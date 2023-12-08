package main

import (
	"log"
	
	"github.com/yurajp/ztube/config"
	"github.com/yurajp/ztube/web"
)

func main() {
	err := config.Prepare()
	if err != nil {
		log.Fatal(err)
	}
	web.WebStart()
	
	Work:
	for {
		select {
			case <-web.ChanQuit:
			  break Work
			default:
		}
	}
}
