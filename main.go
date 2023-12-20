package main

import (
	
	"github.com/yurajp/ztube/web"
	"github.com/yurajp/ztube/player"
)


func main() {

	web.WebStart()
	
	defer func() {
		if player.Playing {
		  player.StopPlay()
		}
	}()
	
	Work:
	for {
		select {
			case <-web.ChanQuit:
			  break Work
			default:
		}
	}
}
