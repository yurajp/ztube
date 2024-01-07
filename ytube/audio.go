package ytube

import (
	"fmt"
	"strings"
	"log"
	"image"
	"bytes"
	"path/filepath"
	"os/exec"

	"github.com/yurajp/ztube/config"
	"github.com/yurajp/ztube/atag"
	"github.com/yurajp/ztube/deezer"
	"github.com/yurajp/ztube/database"
  "github.com/gabriel-vasile/mimetype"
  "github.com/frolovo22/tag"
  ffmpeg "github.com/u2takey/ffmpeg-go"
	"github.com/disintegration/imaging"
	"github.com/otiai10/copy"
)


func audioName() string {
	if Video == "" {
		return ""
	}
	mtyp, _ := mimetype.DetectFile(Video)
	vidext := mtyp.Extension()
	audext := ".mp3"
	return strings.TrimSuffix(Video, vidext) + audext
}



func (o *Opts) MakeAudio() error {
	aname := audioName()
	cmd := exec.Command("ffmpeg", "-i", Video, aname)
	_, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("ConvertingAudioError: %s", err)
	}
	fmt.Println("  Audio extracted")
	searchSng := fmt.Sprintf("%s %s", o.Artist, o.Title)
	dzTr, err := deezer.DeezerTrack(searchSng)
	if err != nil {
		return fmt.Errorf("Deezer error: %s", err)
	}
	o.Duration = dzTr.Duration
	if o.Album == "" {
		o.Album = dzTr.Album.Title
	}
	if !o.Identify {
		atr, err := atag.Recognize(aname)
		if err != nil {
			fmt.Println(err)
		}
		o.Year = atr.Year
		if o.Year == 0 {
			fmt.Println("  no year")
		}
		if atr.Title != o.Title {
			fmt.Printf("  Differents track's titles!\n  1. '%s'\n  2. '%s'\n", atr.Title, o.Title)
		}
	}
	dbTr := database.NewTrack()
	dbTr.Artist = o.Artist
	dbTr.Title = o.Title
	dbTr.Album = o.Album
	dbTr.Year = o.Year
	dbTr.Duration = o.Duration
	dbTr.Code = o.Code
	dbTr.Path = aname
	err = o.SetTags()
	if err != nil {
     return fmt.Errorf("TagsError: %s", err)
	}
	fmt.Println("  Tags setted")
	
	imgName := fmt.Sprintf("%s -%s.jpg", o.Artist, o.Title)
	imgPath := filepath.Join(config.Conf.ImgDir, imgName)
	err = deezer.GetCover(dzTr)
	if err != nil {
		fmt.Println(err)
	} else {
		dbTr.Cover = imgPath
	}
	
	err = dbTr.AddTrackToDb()
	if err != nil {
		fmt.Println(err)
	}
	
	shDir := config.Conf.ShareDir
	if shDir != "" {
		src := filepath.Base(aname)
		shName := filepath.Join(shDir, src)
		err = copy.Copy(aname, shName)
		if err != nil {
			log.Printf("Cannot copy to shareDir: %s\n", err)
		}
	}
	return nil
}


func FrameImage(fileName string, frameNum int) (image.Image, error) {
  ext := filepath.Ext(fileName)
  if ext == ".webm" {
  	return nil, nil
  }
  vcodec := "mjpeg"
  buf := bytes.NewBuffer(nil)
  err := ffmpeg.Input(fileName).
  	Filter("select", ffmpeg.Args{fmt.Sprintf("gte(n,%d)", frameNum)}).
  	Output("pipe:", ffmpeg.KwArgs{"vframes": 1, "format": "image2", "vcodec": vcodec}).
     WithOutput(buf).Run()
	if err != nil {
		return nil, err
	}
  img, err := imaging.Decode(buf)
  if err != nil {
    return nil, err
  }
  return img, nil
}

func ImagePath() string {
	if Video == "" {
		return ""
	}
	prts := strings.Split(Video, ".")
	dir, file := filepath.Split(strings.TrimSuffix(Video, prts[len(prts) - 1]) + "jpg")  // png
	return filepath.Join(dir, "pics", file)
}

func SaveImage(frame, size int) error {
  img, err := FrameImage(Video, frame)
  if err != nil {
    return err
  }
  if img == nil {
  	log.Println("webm - no image")
  	return nil
  }
  img = Crop(img, size)
  imgFile := ImagePath()
  err = imaging.Save(img, imgFile)
  if err != nil {
    return err
  }
  fmt.Println("  DONE")
  
  return nil
}  

func Crop(img image.Image, size int) image.Image {
	hi := img.Bounds().Dy()
	img = imaging.CropAnchor(img, hi, hi, imaging.Center)
	img = imaging.Resize(img, size, 0, imaging.Lanczos)
	return img
}


func (o *Opts) SetTags()	error {
	audio := audioName()
	if filepath.Ext(audio) == ".ogg" {
		return nil
	}
  meta, err := tag.ReadFile(audio)
  if err != nil {
	  return fmt.Errorf("ReadFileError: %s", err)
  }
  err = meta.SetTitle(o.Title)
  if err != nil {
	  return fmt.Errorf("SetTitleTagError: %s", err)
  }
  err = meta.SetArtist(o.Artist)
    if err != nil {
	    return fmt.Errorf("SetArtistTagError: %s", err)
  }
  err = meta.SetAlbum(o.Album)
    if err != nil {
	    return fmt.Errorf("SetAlbumTagError: %s", err)
  }
  err = meta.SetYear(o.Year)
    if err != nil {
	    return fmt.Errorf("SetYearTagError: %s", err)
  }
  
  err = meta.SaveFile(audio)
  if err != nil {
	  return fmt.Errorf("SaveFileError: %s", err)
  }
  
  return nil
}
