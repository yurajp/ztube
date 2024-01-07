package main

import (
  "fmt"
  "github.com/yurajp/ztube/database"
	"github.com/yurajp/ztube/web"
	"github.com/yurajp/ztube/player"
)


func main() {
	
	err := database.RepairImgsDb()
	if err != nil {
		fmt.Println(err)
		return
	}
	
	web.WebStart()
	
	defer func() {
		if player.Playing {
		  player.Current.Stop()
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
