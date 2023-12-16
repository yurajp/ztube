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
	tubeTmp *template.Template
	linkTmp *template.Template
	resTmp *template.Template
  listTmp *template.Template
  songTmp *template.Template
  ChanQuit = make(chan struct{}, 1)
)


func WebStart() {
	addr = ":" + config.Conf.Port
	mux := http.NewServeMux()
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/tube",tubeHandler)
	mux.HandleFunc("/link", linkHandler)
	mux.HandleFunc("/result", resHandler)
	mux.HandleFunc("/watch", watchHandler)
	mux.HandleFunc("/song", songHandler)
	mux.HandleFunc("/action", actionHandler)
	mux.HandleFunc("/quit", quitHandler)
  mux.HandleFunc("/list", listHandler)
	stdir := config.Conf.AppDir + "/web/static"
	fs := http.FileServer(http.Dir(stdir))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	tmpdir := config.Conf.AppDir + "/web/templates/"
	homeTmp, _ = template.ParseFiles(tmpdir + "home.html")
	tubeTmp, _ = template.ParseFiles(tmpdir + "tube.html")
  linkTmp, _ = template.ParseFiles(tmpdir + "link.html")
  resTmp, _ = template.ParseFiles(tmpdir + "result.html")
  songTmp, _ = template.ParseFiles(tmpdir + "song.html")
  listTmp, _ = template.ParseFiles(tmpdir + "list.html")
  
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
