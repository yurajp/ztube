package ytube

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
	Artist string
	Title string
	Code string
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

func (o *Opts) SetCodeFromLink(url string) bool {
	sep := "?v="
	ss := strings.Split(url, sep)
	if len(ss) == 2 {
		o.Code = ss[1]
		return true
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
	o.Code = id
	return id != ""
}

func (o *Opts) SearchQuery() string {
	sng := fmt.Sprintf("%s %s ", o.Artist, o.Title)
  return fmt.Sprintf("https://m.youtube.com/results?sp=mAEA&search_query=%s", url.QueryEscape(sng))
}

func (o *Opts) GetBody() (*[]byte, error) {
	addr := o.SearchQuery()
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

func (o *Opts) SetId(bs *[]byte) bool {
	sep := []byte("\\u0026")
	sl := bytes.Split(*bs, sep)
  f := []byte("watch?v=")
	for _, l := range sl {
		if bytes.Contains(l, f) {
			id := bytes.Split(l, f)[1]
			o.Code = string(id)
			return true
		}
	}
	return o.Code != ""
}

func (o *Opts) MakeCode() error {
	bs, err := o.GetBody()
	if err != nil {
		return fmt.Errorf("Cannot get response: %s", err)
	}
	ok := o.SetId(bs)
	if !ok {
		return errors.New("SetCodeError")
	}
	return nil
}

func (o *Opts) WatchLink() (string, bool) {
	if o.Code != "" {
	  return fmt.Sprintf("https://youtu.be/watch?v=%s", o.Code), true
	}
	return "", false
}

func (o *Opts) FName() string {
	return fmt.Sprintf("%s - %s", o.Artist, o.Title)
}


func (o *Opts) DownloadVideo() error {
	if o.Code == "" {
		return errors.New("EmptyCodeError")
	}
  client := youtube.Client{}
  video, err := client.GetVideo(o.Code)
  if err != nil {
  	return fmt.Errorf("GetVideoErr: %s", err)
  }
  formats := video.Formats.WithAudioChannels() 
	stream, _, err := client.GetStream(video, &formats[0])
	if err != nil {
		return fmt.Errorf("GetStreamErr: %s", err)
	}
	defer stream.Close()

	fname := o.FName()
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

func (o *Opts) OpenSearch() {
  lk := o.SearchQuery()
  exec.Command("xdg-open", lk)
}

func NewOpts() *Opts {
	return &Opts{}
}

func (o *Opts) Produce(lnk string) error {
	path = config.Conf.DirPath
	OnAir = true
	GetCodeError := errors.New("Cannot get code from url")
	
	if o.Manualy {
		o.OpenSearch()
		
		ol := <-LinkChan
		ok := o.SetCodeFromLink(ol)
		if !ok {
			return GetCodeError
		}
	}
	if lnk != "" {
		ok := o.SetCodeFromLink(lnk)
		if !ok {
			return GetCodeError
		}
	}
	if !o.Manualy {
		err := o.MakeCode()
		if err != nil {
			return err
		}
	}
	
	err := o.DownloadVideo()
	if err != nil {
		return err
	}
	if !o.VideoOnly {
	  err = o.MakeAudio()
	  if err != nil {
	  	return err
	  }
	  err = SaveImage(64, 500)
	  if err != nil {
	  	return err
	  }
	}
	if o.AudioOnly {
		os.Remove(Video)
		Video = ""
	}
  OnAir = false
  
	return nil
}
