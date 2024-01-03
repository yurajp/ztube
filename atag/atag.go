package atag

import (
	"net/http"
	"net/url"
	"fmt"
	"strings"
	"path/filepath"
	"encoding/json"
	"io/ioutil"
	"bytes"
	"os"
	"os/exec"
	"io"
  "mime/multipart"
  "time"
  "errors"
  
  "github.com/yurajp/ztube/config"
  "github.com/tcolgate/mp3"
)

var (
	host = "https://audiotag.info/api"
	apikey = "8fd40b9d9fa0473ee3ae317995ac752a"
	Dir = config.Conf.DirPath
	record = "termux-microphone-record"
	RecDir = filepath.Join(config.Conf.AppDir, "rec/")
	RecFile = filepath.Join(RecDir, "current.m4a")
	WavDir = filepath.Join(config.Conf.AppDir, "wavs/")
)

type IdentReply struct {
	Success bool `json:"success"`
	Token string `json:"token"`
	Error string `json:"error"`
	JobStatus string `json:"job_status"`
	StartTime int `json:"start_time"`
	TimeLen int `json:"time_len"`
}

type Track struct {
	Title string
	Artist string
	Album string
	Year int
	Length int
}


type Candidate struct {
	Tracks [][]interface{} `json:"tracks"`
	Confidence int `json:"confidence"`
	Time string `json:"time"`
}

type Result struct {
	Success bool `json:"success"`
	Error string `json:"error"`
	Result string `json:"result"`
	Data []Candidate `json:"data"`
}

type Account struct {
	CreditBalance int `json:"current_credit_balance"`
	SecRemaind int `json:"identification_free_sec_remainder"`
	ExpDate string `json:"expiration_date"`
	QueryCount int `json:"queries_count"`
	UpldSeconds int `json:"uploaded_duration_sec"`
	UpldBytes int `json:"uploaded_size_bytes"`
	CreditSpent int `json:"credits_spent"`
  Success bool `json:"success"`
}

type RecOpts struct {
	F string
	L string
	E string
	B string
	R string
	C string
}


func (tr *Track) HumanLen() string {
  m := tr.Length / 60
	s := tr.Length % 60
	
	return fmt.Sprintf("%d:%.2d", m, s)
}

func Record() error {
	_, err := os.Stat(RecDir)
	if os.IsNotExist(err) {
		os.Mkdir(RecDir, 0770)
	}
	ro := RecOpts{RecFile, "30", "aac", "128", "44100", "1"}
	cmdRec := exec.Command(record, "-f", ro.F, "-l", ro.L, "-e", ro.E, "-b", ro.B, "-r", ro.R, "-c", ro.C)
	cmdQuit := exec.Command(record, "-q")
	
	err = cmdRec.Run()
	if err != nil {
	  return fmt.Errorf("Record error: %s", err)
	}
	time.Sleep(time.Second * 30)
	cmdQuit.Run()
	
	return nil
}

func Remaind(sec int) string {
	h := sec / 3600
	m := (sec / 60) % 60
	s := sec % 60
	
	return fmt.Sprintf("%d:%.2d:%.2d", h, m, s)
}

func Mbyte(b int) string {
	m := float64(b) / 1024.0 / 1024.0
	return fmt.Sprintf("%.2f mB", m)
}

func (ac Account) StatFormat() string {
	cb := fmt.Sprintf(" Credit balance: %d\n", ac.CreditBalance)
	sr := fmt.Sprintf(" Time remained: %s\n", Remaind(ac.SecRemaind))
	ed := fmt.Sprintf(" Expiration date: %s\n", strings.Fields(ac.ExpDate)[0])
	qc := fmt.Sprintf(" Query count: %d\n", ac.QueryCount)
	us := fmt.Sprintf(" Uploaded seconds: %d\n", ac.UpldSeconds)
	ub := fmt.Sprintf(" Uploaded bytes: %s\n", Mbyte(ac.UpldBytes))
	cs := fmt.Sprintf(" Credits spent: %d\n", ac.CreditSpent)
	
	return fmt.Sprintf("%s%s%s%s%s%s%s", cb, sr, ed, qc, us, ub, cs)
}

func GetStatistic() (string, error) {
   	data := url.Values{
   		"action": {"stat"},
   		"apikey": {apikey},
   	}
   	resp, err := http.PostForm(host, data)
   	if err != nil {
   		return "", err
   	}
   	defer resp.Body.Close()
   	body, err := ioutil.ReadAll(resp.Body)
   	if err != nil {
   		return "", err
   	}
   	var acc Account
   	err = json.Unmarshal(body, &acc)
   	if err != nil {
   		return "", err
   	}
   	stat := acc.StatFormat()
   	
   	return stat, nil
}

func shortTitle(t string) string {
	if t == "" {
		return t
	}
	ss := strings.Split(t, " (")
	return ss[0]
}

func (c Candidate) ToTrack() *Track {
	t := c.Tracks[0]
	tr := Track{}
	tr.Title = shortTitle(t[0].(string))
	tr.Artist = t[1].(string)
	tr.Album = t[2].(string)
	tr.Year = int(t[3].(float64))

	return &tr
}

func (r Result) Winner() *Track {
	if len(r.Data) == 0 {
		return &Track{}
	}
	win, n := 0, 0
	for i, c := range r.Data {
		if c.Confidence > win {
			n = i
			win = c.Confidence
		}
	}
	best := r.Data[n]
	
	return best.ToTrack()
}

func InfoForm() (string, error) {
	data := url.Values{
		"action": {"info"},
		"apikey": {apikey},
	}
	resp, err := http.PostForm(host, data)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func ConvertToWav(dir, f string) (func(), error) {
	out := Wave(f)
	path := filepath.Join(dir, f)
	cmd := exec.Command("ffmpeg", "-i", path, "-ar", "8000", "-ac", "1", "-vn", out)
	
	err := cmd.Run()
	if err != nil {
		return func(){}, err
	}
	clear := func() {
		os.Remove(out)
	}
	return clear, nil
}

func Wave(f string) string {
	ext := filepath.Ext(f)
	name := filepath.Base(f)
	return filepath.Join(WavDir, strings.TrimSuffix(name, ext) + ".wav")
}

func SendFileToIdent(name string) (string, string, error) {
  buf := new(bytes.Buffer)
  w := multipart.NewWriter(buf)
  wavName := filepath.Base(Wave(name))
	upl, err := w.CreateFormFile("file", wavName)
	if err != nil {
		return "", "", err
	}
	bs, err := os.ReadFile(Wave(name))
	if err != nil {
		return "", "", err
	}
	_, err = upl.Write(bs)
	if err != nil {
		return "", "", err
	}
	actWr, err := w.CreateFormField("action")
	if err != nil {
		return "", "", err
	}
	_, err = actWr.Write([]byte("identify"))
	if err != nil {
		return "", "", err
	}
	keyWr, err := w.CreateFormField("apikey")
	if err != nil {
		return "", "", err
	}
	_, err = keyWr.Write([]byte(apikey))
	if err != nil {
		return "", "", err
	}
	err = w.Close()
	if err != nil {
		return "", "", err
	}
	
  req, err := http.NewRequest("POST", host, buf)
  if err != nil {
    return "", "", err
  }
  req.Header.Add("Content-Type", w.FormDataContentType())
  
  client := &http.Client{}
  res, err := client.Do(req)
  if err != nil {
    return "", "", err
  }
  defer res.Body.Close()
  body, err := io.ReadAll(res.Body)
  if err != nil {
    return "", "", err
  }
  var ir IdentReply
  err = json.Unmarshal(body, &ir)
  if err != nil {
	  return "", "", err
  }
  
  return ir.Token, ir.Error, nil
}

func GetIdentResult(token string) (Result, error) {
	data := url.Values{
		"action": {"get_result"},
		"token": {token},
		"apikey": {apikey},
	}
	n := 0
	for n < 50 {
	  resp, err := http.PostForm(host, data)
	  if err != nil {
		  return Result{}, err
	  }
	  defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
	  if err != nil {
		  return Result{}, err
	  }
	  var res Result
	  err = json.Unmarshal(body, &res)
	  if err != nil {
		  return Result{}, err
	  }
	  if res.Result != "wait" {
		  return res, nil
	  }
	  n++
	  time.Sleep(time.Second)
	}
	return Result{}, nil
}

func GetDuration(sng string) int {
	fl, err := os.Open(sng)
	if err != nil {
		fmt.Println(err)
		return 0
	}
	defer fl.Close()
	
	d := mp3.NewDecoder(fl)
  var f mp3.Frame
  skipped := 0
  t := 0.0
  for {
    if err := d.Decode(&f, &skipped); err != nil {
      if err == io.EOF {
        break
      }
        fmt.Println(err)
          return 0
    }
    t += f.Duration().Seconds()
  }
  return int(t)
}


func Recognize(file string) (*Track, error) {
	var err error
  dir := ""
	if file == "" {
		file = RecFile
		dir = RecDir
		err = Record()
		if err != nil {
			return &Track{}, fmt.Errorf("Record error: %s", err)
		}
	} else {
		d := filepath.Dir(file)
		if d == "." {
		  dir = Dir
		} else {
			dir = d
		}
	}
  nm := filepath.Base(file)
	clr, err := ConvertToWav(dir, nm)
	if err != nil {
		return &Track{}, fmt.Errorf("convert to wav: %s", err)
	}
	defer clr()
	
	token, serr, err := SendFileToIdent(file)
	if err != nil {
		return &Track{}, fmt.Errorf("SendFile: %s", err)
	}
  if token == "" {
	  return &Track{}, fmt.Errorf("reply on SendFile: %s", serr)
  }
  res, err := GetIdentResult(token)
  if err != nil {
	  return &Track{}, fmt.Errorf("GetIdentResult: %s", err)
  }
  if !res.Success {
	  return &Track{}, fmt.Errorf("Bad result: %s - %s", res.Result, res.Error)
  }
  if res.Result == "not found" {
	  return &Track{}, errors.New("Track Not Found")
  }
  track := res.Winner()
  
  return track, nil
}

func SecToMin(sec int) string {
	m := sec / 60
	s := sec % 60
	return fmt.Sprintf("%d:%.2d", m, s)
}

func (t *Track) String() string {
	tt := "Title: " + t.Title
	ar := "Artist: " + t.Artist
	al := "Album: " + t.Album
	ye := fmt.Sprintf("Year: %d", t.Year)
	ms := t.HumanLen()
	ln := "Length: " + ms
	
	return fmt.Sprintf(" %s\n %s\n %s\n %s\n %s\n", tt, ar, al, ye, ln)
}

func (tr *Track) Mp3Name() string {
	return fmt.Sprintf("%s - %s.mp3", tr.Artist, tr.Title)
}
