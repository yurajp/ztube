package web


import (
	"net/http"
	"html/template"
	"log"
	"fmt"
	"os/exec"
	"path/filepath"
	"os"
	
	"github.com/yurajp/ztube/config"
)

var (
	homeTmp *template.Template
	tubeTmp *template.Template
	linkTmp *template.Template
	resTmp *template.Template
  listTmp *template.Template
  songTmp *template.Template
  ChanQuit = make(chan struct{}, 1)
)


func WebStart() {
	addr := ":" + config.Conf.Port
	mux := http.NewServeMux()
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/recognize", recognizeHandler)
	mux.HandleFunc("/tube",tubeHandler)
	mux.HandleFunc("/link", linkHandler)
	mux.HandleFunc("/result", resHandler)
	mux.HandleFunc("/watch", watchHandler)
	mux.HandleFunc("/song", songHandler)
	mux.HandleFunc("/action", actionHandler)
	mux.HandleFunc("/quit", quitHandler)
  mux.HandleFunc("/list", listHandler)
  mux.HandleFunc("/rand", randHandler)
	stdir := config.Conf.AppDir + "/web/static"
	fs := http.FileServer(http.Dir(stdir))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	tmpdir := config.Conf.AppDir + "/web/templates/"
	homeTmp, _ = template.ParseFiles(tmpdir + "home.html")
	tubeTmp, _ = template.ParseFiles(tmpdir + "tube.html")
  linkTmp, _ = template.ParseFiles(tmpdir + "link.html")
  resTmp, _ = template.ParseFiles(tmpdir + "result.html")
  songTmp, _ = template.ParseFiles(tmpdir + "song.html")
  lTmp, err := template.ParseFiles(tmpdir + "list.html")
  if err != nil {
  	log.Printf(" Parse listTmp: %s\n", err)
  } else {
  	listTmp = lTmp
  }
  
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
	imgPath := filepath.Join(config.Conf.AppDir, "web", "static", "currImg")
	os.Create(imgPath)
	exec.Command("xdg-open", "http://localhost" + addr).Run()
	
	fmt.Printf("\n    ZTUBE\n  Server runing on %s\n", addr)
}

