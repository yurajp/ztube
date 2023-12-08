package utube

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"io"
	"os"
	"os/exec"
	"bytes"
	"errors"
	
	"github.com/yurajp/ztube/config"
  "github.com/kkdai/youtube/v2"
  "github.com/gabriel-vasile/mimetype"
)

type Opts struct {
	Song string
	Link string
	Manualy bool
	VideoOnly bool
	AudioOnly bool
}

var (
	path string
	LinkChan = make(chan string, 1)
	OnAir = false
	Video string
)

func GetIdFromLink(url string) string {
	sep := "?v="
	ss := strings.Split(url, sep)
	if len(ss) == 2 {
		return ss[1]
	}
	id := ""
	px := "https://youtu.be/"
	if strings.HasPrefix(url, px) {
		id = strings.TrimPrefix(url, px)
	}
	sx := "?feature=shared"
	if strings.HasSuffix(id, sx) {
		id = strings.TrimSuffix(id, sx)
	}
	return id
}

func GetBody(sng string) (*[]byte, error) {
	addr := fmt.Sprintf("https://m.youtube.com/results?sp=mAEA&search_query=%s", url.QueryEscape(sng))
	resp, err := http.Get(addr)
	if err != nil {
		return &[]byte{}, err
	}
  defer resp.Body.Close()
  if resp.StatusCode != 200 {
    return &[]byte{}, errors.New(fmt.Sprintf("%d %s", resp.StatusCode, resp.Status))
  } 
  buf, err := io.ReadAll(resp.Body)
  if err != nil {
	  return &[]byte{}, err
  }
  return &buf, nil
}

func FindId(bs *[]byte) string {
	sep := []byte("\\u0026")
	sl := bytes.Split(*bs, sep)
  f := []byte("watch?v=")
	for _, l := range sl {
		if bytes.Contains(l, f) {
			id := bytes.Split(l, f)[1]
			return string(id)
		}
	}
	return ""
}

func MakeId(sng string) (string, error) {
	bs, err := GetBody(sng)
	if err != nil {
		return "", fmt.Errorf("Cannot get response: %s", err)
	}
	id := FindId(bs)
	if id == "" {
		return "", fmt.Errorf("Cannot find id: %s", err)
	}
	fmt.Printf("  Code: %s\n", id)
	
	return id, nil
}

func MakeLink(id string) string {
	return fmt.Sprintf("https://youtu.be/watch?v=%s", id)
}

func DownloadVideo(id, name string) error {
	if id == "" {
		return errors.New("EmptyIdError")
	}
  client := youtube.Client{}
  video, err := client.GetVideo(id)
  if err != nil {
  	return fmt.Errorf("GetVideoErr: %s", err)
  }
  formats := video.Formats.WithAudioChannels() 
	stream, _, err := client.GetStream(video, &formats[0])
	if err != nil {
		return fmt.Errorf("GetStreamErr: %s", err)
	}
	defer stream.Close()

	fname := Name2Song(name)
	fpath := path + fname
	file, err := os.Create(fpath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, stream)
	if err != nil {
		return err
	}
	
  mtype, err := mimetype.DetectFile(fpath)
  if err != nil {
  		return fmt.Errorf("Cannot detect MIME: %s", err)
  }
  ext := mtype.Extension()
  Video = fpath + ext
  
  err = os.Rename(fpath, Video)
	if err != nil {
		return fmt.Errorf("Cannot rename: %s", err)
	}
	fmt.Printf("\n  Downloaded  %s\n  ", fname + ext)
  
	return nil
}


func Name2Song(sng string) string {
	return strings.ReplaceAll(sng, " ", "_")
}

func ConvertToAudio() {
	mtyp, err := mimetype.DetectFile(Video)
	if err != nil {
		fmt.Println(err)
		return
	}
	audio := strings.TrimSuffix(Video, mtyp.Extension()) + ".mp3"
	cmd := exec.Command("ffmpeg", "-i", Video, audio)
	_, _ = cmd.Output()
	
	fmt.Println("  Audio extracted")
}

func OpenSearch(name string) {
	sng := url.QueryEscape(name)
  lk :=  "https://m.youtube.com/results?sp=mAEA&search_query=" + sng
  exec.Command("xdg-open", lk)
}


func (o Opts) Produce() error {
	path = config.Conf.DirPath
	id := ""
	OnAir = true
	
	if o.Manualy && o.Link == "" {
		OpenSearch(o.Song)
		
		ol := <-LinkChan
		o.Link = ol
	}
	if o.Manualy && o.Link != "" {
		id = GetIdFromLink(o.Link)
	}
	if !o.Manualy {
		i, err := MakeId(o.Song)
		if err != nil {
			return err
		}
		id = i
	}
	name := Name2Song(o.Song)
	err := DownloadVideo(id, name)
	if err != nil {
		return err
	}
	if !o.VideoOnly {
	  ConvertToAudio()
	}
	if o.AudioOnly {
		os.Remove(Video)
		Video = ""
	}
  OnAir = false
  
	return nil
}
