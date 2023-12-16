package web

import (
	"net/http"
	"net/url"
  "log"
  "os/exec"
  "fmt"
  
  "github.com/yurajp/ztube/ytube"
  "github.com/yurajp/ztube/player"
  "github.com/yurajp/ztube/config"
  
)

var (
	Dir = config.Conf.DirPath
	Addr = ":" + config.Conf.Port + "/"
	PList = player.PList
//	Current *player.Song
	Status string
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	err := homeTmp.Execute(w, Addr)
	if err != nil {
		log.Printf(" homeTemplateErr: %s\n", err)
		return
	}
}

func tubeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tubeTmp.Execute(w, nil)
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
		return
	}
	exec.Command("xdg-open", v).Run()
}

//
func listHandler(w http.ResponseWriter, r *http.Request) {
  if PList.Dir == "" {
    pl, err := player.Mp3List(Dir)
    if err != nil {
  	  http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    PList = pl
  }
  listTmp.Execute(w, PList)
}


func songHandler(w http.ResponseWriter, r *http.Request) {
  path := r.URL.Query().Get("path")
  path, _ = url.QueryUnescape(path)
  if path != "" {
    cur, err := player.MakeSong(path)
    if err != nil {
    	fmt.Println("MakeSongError: ", err)
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
    err = cur.CurPicture()
    if err != nil {
    	fmt.Println("CurPicture", err)
    }
    player.Current = cur
  }
  
  songTmp.Execute(w, player.Current)
 }
 
 
 func actionHandler(w http.ResponseWriter, r *http.Request) {
  act := r.URL.Query().Get("act")
  Status = player.Current.Info()
  switch (act) {
	  case "play":
	    if Status == "Stopped" {
	      player.Current.Play()
	      fmt.Printf("\n >>> %s", player.Current.Title)
	    }
		  if Status == "Paused" {
		  	player.Current.Resume()
		  	fmt.Print(" >>> resumed")
		  }
	  case "pause":
	    if Status == "Playing" {
	      player.Current.Pause()
	      fmt.Print(" >> paused")
	    }
	  case "stop":
	    if Status != "Stopped" {
	      player.Current.Stop()
	      fmt.Println(" > stopped")
	    }
	  default:
  }
  Status = player.Current.Info()
 
  songTmp.Execute(w, player.Current)
}
