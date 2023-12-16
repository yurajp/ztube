package main

import (
	
	"github.com/yurajp/ztube/web"
)


func main() {

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
