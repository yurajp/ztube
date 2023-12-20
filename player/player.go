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
)

var (
	Dir = config.Conf.DirPath
	Addr = ":" + config.Conf.Port
	Current *Song
	Playing bool
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
	pnm := filepath.Join(Dir, "pics", line + ".png")
	_, err := os.Stat(pnm)
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

func RandPlay(pl *Playlist) error {
	lst := pl.List
	length := len(lst)
	if length == 0 {
		return errors.New("Empty Playlist")
	}
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	rnd.Shuffle(5, func(i, j int){
		lst[i], lst[j] = lst[j], lst[i]
	})

	pl.List = lst
	for _, s := range lst {
		s.Play()
		for {
      stat := s.Info()
		  switch (stat) {
			  case "Stopped":
			    Playing =false
			    break
			  default:
			    time.Sleep(time.Millisecond * 500)
			}
		}
	}
	StopPlay()
	return nil
}

func StopPlay() {
	exec.Command("termux-media-player", "stop").Run()
	Playing = false
}
