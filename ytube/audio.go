package ytube

import (
	"fmt"
	"strings"
	"image"
	"bytes"
	"path/filepath"
	"os/exec"
	
  "github.com/gabriel-vasile/mimetype"
  "github.com/frolovo22/tag"
  ffmpeg "github.com/u2takey/ffmpeg-go"
	"github.com/disintegration/imaging"
)

func audioName() string {
	if Video == "" {
		return ""
	}
	prts := strings.Split(Video, ".")
	return strings.TrimSuffix(Video, prts[len(prts) - 1]) + "mp3"
}

func (o *Opts) MakeAudio() error {
	mtyp, err := mimetype.DetectFile(Video)
	if err != nil {
		return fmt.Errorf("MimeError: %s", err)
	}
	audio := strings.TrimSuffix(Video, mtyp.Extension()) + ".mp3"
	cmd := exec.Command("ffmpeg", "-i", Video, audio)
	_, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("ConvertingAudioError: %s", err)
	}
	fmt.Println("  Audio extracted")
	
	err = o.SetTags()
	if err != nil {
     return fmt.Errorf("TagsError: %s", err)
	}
	fmt.Println("  Tags setted")
	return nil
}

func FrameImage(fileName string, frameNum int) (image.Image, error) {
  buf := bytes.NewBuffer(nil)
  err := ffmpeg.Input(fileName).
  	Filter("select", ffmpeg.Args{fmt.Sprintf("gte(n,%d)", frameNum)}).
  	Output("pipe:", ffmpeg.KwArgs{"vframes": 1, "format": "image2", "vcodec": "mjpeg"}).
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

func ImageName() string {
	if Video == "" {
		return ""
	}
	prts := strings.Split(Video, ".")
	dir, file := filepath.Split(strings.TrimSuffix(Video, prts[len(prts) - 1]) + "png")
	return filepath.Join(dir, "pics", file)
}

func SaveImage(frame, size int) error {
  img, err := FrameImage(Video, frame)
  if err != nil {
    return err
  }
  img = Crop(img, size)
  imgFile := ImageName()
  err = imaging.Save(img, imgFile)
  if err != nil {
    return err
  }
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
  
  err = meta.SaveFile(audio)
  if err != nil {
	  return fmt.Errorf("SaveFileError: %s", err)
  }
  
  return nil
}
