package player

import (
	"os"
	"fmt"
	"path/filepath"
	"strings"
	"net/url"
	"os/exec"
	"errors"
	"math/rand"
	"time"
	
	"github.com/yurajp/ztube/config"
  "github.com/frolovo22/tag"
)

var (
	Dir = config.Conf.DirPath
	Addr = ":" + config.Conf.Port
  StopCh = make(chan struct{}, 1)
	Current *Song
	Playing bool
	RandCh = make(chan string, 1)
)

type Playlist struct {
	Dir string
	List []*Song
}
type Song struct {
	Path string
	Artist string
	Title string
	Picture string
}

func (pl *Playlist) Play() bool {
	return Playing
}

func Mp3List(dir string) (*Playlist, error) {
	ins, err := os.ReadDir(dir)
	if err != nil {
		return &Playlist{}, err
	}
	lst := []*Song{}
	for _, f := range ins {
		nm := f.Name()
		ext := filepath.Ext(nm)
		if ext == ".mp3" || ext == ".ogg" {
			s, err := MakeSong(nm)
			if err != nil {
				fmt.Println(err)
				return &Playlist{}, err
			}
		  lst = append(lst, s)
		}
	}
	return &Playlist{dir, lst}, nil
}

func MakeSong(path string) (*Song, error) {
	if path == "" {
		return &Song{}, errors.New("No path passed")
	}
	name := filepath.Base(path)
	line := strings.TrimSuffix(name, filepath.Ext(name))
	fds := strings.Split(line, " - ")
	art, ttl := fds[0], fds[1]
	pnm := filepath.Join(Dir, "pics", line + ".jpg")
	_, err := os.Stat(pnm)
	if os.IsNotExist(err) {
		pnm = filepath.Join(Dir, "pics", line + ".png")
	}
	_, err = os.Stat(pnm)
	if os.IsNotExist(err) {
		pnm = filepath.Join(config.Conf.AppDir, "web", "static", "Z2.png")
	}
	fpath := filepath.Join(Dir, name)
	
	return &Song{fpath, art, ttl, pnm}, nil
}

func (s *Song) Query() string {
	return url.QueryEscape(s.Path)
}

func (s *Song) Href() string {
	qp := s.Query()
	return fmt.Sprintf("http://localhost%s/song?path=%s", Addr, qp)
}

func (s *Song) HostLink() string {
	return fmt.Sprintf("http://localhost%s/", Addr)
}

func (s *Song) Album() string {
  meta, err := tag.ReadFile(s.Path)
  if err != nil {
	  return ""
  }
  alb, err := meta.GetAlbum()
  if err != nil {
  	return ""
  }
  return alb
}

func (s *Song) Year() string {
  meta, err := tag.ReadFile(s.Path)
  if err != nil {
	  return ""
  }
  year, err := meta.GetYear()
  if err != nil {
  	return ""
  }
  return fmt.Sprintf("%d", year)
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

func (s *Song) CurPicture() error {
	bts, err := os.ReadFile(s.Picture)
	if err != nil {
		return err
	}
	apdir := config.Conf.AppDir
	err = os.WriteFile(filepath.Join(apdir,  "web/static/current.png"), bts, 0640)
	if err != nil {
		return err
	}
	return nil
}

func (s *Song) IsCurr() bool {
	return s == Current
}

func RandPlay(pl *Playlist) error {
	if pl == nil {
		p, err := Mp3List(Dir)
		if err != nil {
			return err
		}
		pl = p
	}
	lst := pl.List
	length := len(lst)
	if length == 0 {
		return errors.New("Empty Playlist")
	}
  rand.Seed(time.Now().UnixNano())
	rand.Shuffle(length, func(i, j int) {
		lst[i], lst[j] = lst[j], lst[i]
	})
	
	go func() {
		Play:
	  for _, s := range lst {
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
