package web

import (
	"net/http"
  "log"
  "strings"
  "os/exec"
  
  "github.com/yurajp/ztube/config"
  "github.com/yurajp/ztube/utube"
)


func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		homeTmp.Execute(w, nil)
	}
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.Printf("ParseFormError: %w", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		name := r.FormValue("name")
		link := r.FormValue("link")
		man := r.FormValue("manualy") == "true"
		von := r.FormValue("vonly") == "true"
		aon := r.FormValue("aonly") == "true"
		opts := utube.Opts{name, link, man, von, aon}
		
		go func() {
			err = opts.Produce()
		  if err != nil {
		  	log.Printf("ProduceError: %w", err)
		  }
		}()
		
		if man && link == "" {
			http.Redirect(w, r, "/link", 302)
		} else {
			http.Redirect(w, r, "/result", 302)
		}
	}
}

func linkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		linkTmp.Execute(w, nil)
	}
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.Printf("ParseLinkFormError: %s", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		lk := r.FormValue("link")
		utube.LinkChan <- lk
		http.Redirect(w, r, "/result", 302)
	}
}

func resHandler(w http.ResponseWriter, r *http.Request) {
	wait := utube.OnAir
	if wait {
    resTmp.Execute(w, "WAIT...")
	} else {
    resTmp.Execute(w, "SUCCESS!")
	}
}

func quitHandler(w http.ResponseWriter, r *http.Request) {
	oa := utube.OnAir
	if !oa {
	  resTmp.Execute(w, "CLOSED")
	  ChanQuit <- struct{}{}
	}
}

func contHandler(w http.ResponseWriter, r *http.Request) {
	dir := config.Conf.DirPath
	dir = strings.TrimSuffix(dir, "/")
	exec.Command("nnn", dir).Run()
	http.Redirect(w, r, "/", 302)
}
