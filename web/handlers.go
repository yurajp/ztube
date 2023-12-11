package web

import (
	"net/http"
  "log"
  "os/exec"
  
  "github.com/yurajp/ztube/ytube"
)


func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		homeTmp.Execute(w, nil)
	}
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.Printf("ParseFormError: %s", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		art := r.FormValue("artist")
		ttl := r.FormValue("title")
		link := r.FormValue("link")
		man := r.FormValue("manualy") == "true"
		von := r.FormValue("vonly") == "true"
		aon := r.FormValue("aonly") == "true"
		opts := &ytube.Opts{}
		opts.Artist = art
		opts.Title = ttl
		opts.Manualy = man
		opts.VideoOnly = von
		opts.AudioOnly = aon
		if link != "" {
	  	opts.Manualy = false
		}
		
		go func() {
			err = opts.Produce(link)
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
		ytube.LinkChan <- lk
		http.Redirect(w, r, "/result", 302)
	}
}

func resHandler(w http.ResponseWriter, r *http.Request) {
	wait := ytube.OnAir
	if wait {
    resTmp.Execute(w, "WAIT...")
	} else {
    resTmp.Execute(w, "SUCCESS!")
	}
}

func quitHandler(w http.ResponseWriter, r *http.Request) {
	oa := ytube.OnAir
	if !oa {
	  resTmp.Execute(w, "CLOSED")
	  ChanQuit <- struct{}{}
	}
}


func watchHandler(w http.ResponseWriter, r *http.Request) {
	v := ytube.Video
	if v == "" {
		return
	}
	exec.Command("xdg-open", v).Run()
}
