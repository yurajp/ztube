package web


import (
	"net/http"
	"html/template"
	"log"
	"os/exec"
	
	"github.com/yurajp/ztube/config"
)

var (
	addr string
	homeTmp *template.Template
	linkTmp *template.Template
	resTmp *template.Template
  ChanQuit = make(chan struct{}, 1)
)


func WebStart() {
	addr = ":" + config.Conf.Port
	mux := http.NewServeMux()
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/link", linkHandler)
	mux.HandleFunc("/result", resHandler)
	mux.HandleFunc("/quit", quitHandler)
	stdir := config.Conf.AppDir + "/web/static"
	fs := http.FileServer(http.Dir(stdir))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	tmpdir := config.Conf.AppDir + "/web/templates/"
	homeTmp, _ = template.ParseFiles(tmpdir + "home.html")
  linkTmp, _ = template.ParseFiles(tmpdir + "link.html")
  resTmp, _ = template.ParseFiles(tmpdir + "result.html")
  
  server := http.Server{
  	Addr: addr,
  	Handler: mux,
  }
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}()
	
	exec.Command("xdg-open", "http://localhost" + addr).Run()
}
