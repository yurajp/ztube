package player

import (
	"fmt"
	"strings"
	"os/exec"
	"errors"
	"math/rand"
	"time"
	"path/filepath"
	
	"github.com/yurajp/ztube/config"
	"github.com/yurajp/ztube/database"
  "github.com/otiai10/copy"
)

var (
//	Dir = config.Conf.DirPath
	Addr = ":" + config.Conf.Port
  StopCh = make(chan struct{}, 1)
	Current *Song
	Playing bool
	RandCh = make(chan string, 1)
)

type Playlist struct {
	List []*Song
}

type Song database.Track
	

func (pl Playlist) GetSong(id int) *Song {
	for _, s := range pl.List {
		if s.Id == int64(id) {
			return s
		}
	}
	return &Song{}
}

func (pl Playlist) Length() int {
	return len(pl.List)
}

func (pl Playlist) Play() bool {
	return Playing
}

func NewList() (Playlist, error) {
  tracks, err := database.GetAllTracksFromDb()
  if err != nil {
  	return Playlist{}, err
  }
	lst := []*Song{}
	for _, tr := range tracks {
		s := Song(tr)
		lst = append(lst, &s)
	}
	return Playlist{lst}, nil
}

func (s *Song) Href() string {
	return fmt.Sprintf("http://localhost%s/song?id=%d", Addr, s.Id)
}

func (s *Song) HostLink() string {
	return fmt.Sprintf("http://localhost%s/", Addr)
}

func (s *Song) ImgStat() string {
	if s.Cover == "" {
		return "static/Z2.png"
	}
	stFile := filepath.Join(config.Conf.AppDir, "web", "static", "currImg")
	err := copy.Copy(s.Cover, stFile)
	if err != nil {
		fmt.Printf("copy Cover: %s\n", err)
	}
	return "static/currImg"
}


func (s *Song) Play() {
	Current = s
	exec.Command("termux-media-player", "play", s.Path).Run()
	Playing = true
}

func (s *Song) ActPlay() string {
	return fmt.Sprintf("%saction?act=play", s.HostLink())
}

func (s *Song) Pause() {
	exec.Command("termux-media-player", "pause").Run()
}

func (s *Song) ActPause() string {
	return fmt.Sprintf("%saction?act=pause", s.HostLink())
}

func (s *Song) Stop() {
	exec.Command("termux-media-player", "stop").Run()
	Playing = false
}

func (s *Song) ActStop() string {
	return fmt.Sprintf("%saction?act=stop", s.HostLink())
}

func (s *Song) Resume() {
	exec.Command("termux-media-player", "play").Run()
}

func (s *Song) Info() string {
	inf, _ := exec.Command("termux-media-player", "info").Output()
	stat := strings.TrimPrefix(string(inf), "Status: ")
	statword := strings.Fields(stat)[0]
	if statword == "No" {
		return "Stopped"
	}
	return statword
}

func (s *Song) IsCurr() bool {
	return s == Current
}

func RandPlay(pl Playlist) error {
	p, err := NewList()
	if err != nil {
		return err
	}
	pl = p
	
	mpk := pl.List
	length := len(mpk)
	if length == 0 {
		return errors.New("Empty Playlist")
	}
  rand.Seed(time.Now().UnixNano())
	rand.Shuffle(length, func(i, j int) {
		mpk[i], mpk[j] = mpk[j], mpk[i]
	})
	
	go func() {
		Play:
	  for _, s := range mpk {
      if !Playing {
        RandCh <-s.Title
		    Current = s
        s.Play()
        Wait:
		    for {
		      select {
		      case <-StopCh:
		        RandCh <-""
		        fmt.Println("  Interrupted")
		        Current.Stop()
            Playing = false
		        break Play
		      default:
            stat := s.Info()
		        switch (stat) {
			        case "Stopped":
			        Playing = false
			        break Wait
			      default:
			        time.Sleep(time.Millisecond * 300)
		        }
		  	  }
		    }
		  }
	  }
  }()
  
	return nil
}
