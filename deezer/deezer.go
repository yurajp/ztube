package deezer

import (
	"fmt"
	"net/http"
	"encoding/json"
	"strings"
	"path/filepath"
	"io"
	"net/url"
	"os"
	"errors"
	
	"github.com/yurajp/ztube/config"
	"github.com/cavaliergopher/grab/v3"
)

type Deezer struct {
	Data []Track `json:"data"`
}

type Track struct {
	Id int `json:"id"`
	Readable bool `json:"readable"`
	Title string `json:"title"`
	TitleShort string `json:"title_short"`
	TitleVersion string `json:"title_version"`
	ISRC string `json:"isrc"`
	Link string `json:"link"`
	Duration int `json:"duration"`
	TrackPosition int `json:"track_position"`
	DiskNumber int `json:"disk_number"`
	Rank int `json:"rank"`
	Xlyrics bool `json:"explicit_lyrics"`
	XClyrics int `json:"explicit_content_lyrics"`
	XCcover int `json:"explicit_content_cover"`
	Prewiew string `json:"preview"`
	Md5 string `json:"md5_image"`
	Artist `json:"artist"`
	Album `json:"album"`
	Type string `json:"type"`
}

type Artist struct {
	Id int `json:"id"`
	Name string `json:"name"`
	Link string `json:"link"`
	Picture string `json:"picture"`
	PicSmall string `json:"picture_small"`
	PicMed string `json:"picture_medium"`
	PicBig string `json:"picture_big"`
	PicXL string `json:"picture_xl"`
	Tracklist string `json:"tracklist"`
	Type string `json:"type"`
}


type Album struct {
	Id int `json:"id"`
	Title string `json:"title"`
	Cover string `json:"cover"`
	CoverSmall string `json:"cover_small"`
	CoverMed string `json:"cover_medium"`
	CoverBig string `json:"cover_big"`
	CoverXL string `json:"cover_xl"`
	Md5 string `json:"md5_image"`
	Tracklist string `json:"tracklist"`
	Type string `json:"type"`
}


func DeezerTrack(sng string) (*Track, error) {
  qes := url.QueryEscape(sng)
	url := "https://deezerdevs-deezer.p.rapidapi.com/search?q=" + qes

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("X-RapidAPI-Key", "f9b37554dcmsh46f32a259a79273p10764djsn77afdff2674c")
	req.Header.Add("X-RapidAPI-Host", "deezerdevs-deezer.p.rapidapi.com")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return &Track{}, err
	}

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	var dz Deezer
	err = json.Unmarshal(body, &dz)
	if err != nil {
		return &Track{}, err
	}
  tr := dz.Data[0]
  
  return &tr, nil
}

func ShortTitle(t string) string {
	return strings.Split(t, " (")[0]
}

func GetCover(tr *Track) error {
	img := tr.Album.CoverBig
	if img == "" {
		return errors.New("No image for cover")
	}
	dir := config.Conf.ImgDir
	 _, err := os.Stat(dir)
	if os.IsNotExist(err) {
	 	os.Mkdir(dir, 0750)
	}
	resp, err := grab.Get(dir, tr.Album.CoverBig)
	if err != nil {
		return err
	}
	imgName := resp.Filename
	ext := filepath.Ext(imgName)
	newName := fmt.Sprintf("%s - %s%s", tr.Artist.Name, tr.Title, ext)
	newPath := filepath.Join(dir, newName)
	err = os.Rename(imgName, newPath)
	if err != nil {
		return err
	}
	fmt.Println("Cover downloaded")
	return nil
}
