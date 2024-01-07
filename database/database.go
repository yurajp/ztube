package database

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"os"
	
	"github.com/yurajp/ztube/config"
  _ "github.com/mattn/go-sqlite3"
)

type Track struct {
	Id int64
	Artist string
	Title string
	Album string
	Year int
	Duration int
	Code string
	Path string
	Cover string
}

var (
//	DbTrack *Track
	dbPath = filepath.Join(config.Conf.AppDir, "database/songs.db")
)

func init() {
	err := CheckDb()
	if err != nil {
		fmt.Println(err)
	}
}


func NewTrack() *Track {
	return &Track{}
}

func CheckDb() error {
  db, err := sql.Open("sqlite3", dbPath)
  if err != nil {
    return fmt.Errorf("CheckDb error: %s", err)
  }
  defer db.Close()
  create, err := db.Prepare(`CREATE TABLE IF NOT EXISTS songs (id INTEGER PRIMARY KEY AUTOINCREMENT, artist TEXT, title TEXT NOT NULL, album TEXT, year INTEGER, duration INTEGER, code TEXT, path TEXT, cover TEXT)`)
  
  if err != nil {
    return err
  }
  _, err = create.Exec()
  return err
}

func (t *Track) AddTrackToDb() error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()
	
	add, err := db.Prepare(`INSERT INTO songs(artist, title, album, year, duration, code,  path, cover) VALUES (?,?,?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	_, err = add.Exec(t.Artist, t.Title, t.Album, t.Year, t.Duration, t.Code, t.Path, t.Cover)
	if err != nil {
		return err
	}
	return nil
}

func (tr *Track) HumanLen() string {
  m := tr.Duration / 60
	s := tr.Duration % 60
	return fmt.Sprintf("%d:%.2d", m, s)
}



func GetAllTracksFromDb() ([]Track, error) {
  db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return []Track{}, err
	}
	defer db.Close()
	
	query := `SELECT * FROM songs`
	rows, err := db.Query(query)
	if err != nil {
		return []Track{}, err
	}
	tracks := []Track{}
	for rows.Next() {
		var t Track
		err = rows.Scan(&t.Id, &t.Artist, &t.Title, &t.Album, &t.Year, &t.Duration, &t.Code, &t.Path, &t.Cover)
		if err != nil {
			return []Track{}, err
		}
		tracks = append(tracks, t)
	}
	return tracks, nil
}

func (t Track) StringShort() string {
	return fmt.Sprintf(
		"  # %v:\n%s\n%s\n%s\n%d\n%s",
		t.Id, t.Title,
		t.Artist, t.Album,
		t.Year, t.HumanLen())
}

func RemoveTrackFromDb(id int64) error {
  db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()
	
	query, err := db.Prepare(`DELETE FROM songs WHERE id = ?`)
	if err != nil {
		return err
	}
	_, err = query.Exec(id)
	if err != nil {
		return err
	}
	fmt.Println("  row deleted")
	return nil
}

func RepairImgsDb() error {
	tracks, err := GetAllTracksFromDb()
	if err != nil {
		return fmt.Errorf("GetTracks: %s", err)
	}
	dir := config.Conf.ImgDir
	for _, tr := range tracks {
		if tr.Cover != "" {
			continue
		}
		fname := filepath.Base(tr.Path)
		ext := filepath.Ext(fname)
		iname := strings.TrimSuffix(fname, ext) + ".jpg"
		ipath := filepath.Join(dir, iname)
		if _, err = os.Stat(ipath); err != nil {
			continue
		}
		tr.Cover = ipath
		
		err = tr.updateImg()
		if err != nil {
			return fmt.Errorf("Update img: %s", err)
		}
	}
	return nil
}

func (tr Track) updateImg() error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()
	
	upd := `UPDATE songs SET cover = ? WHERE id = ?`
	
	_, err = db.Exec(upd, tr.Cover, tr.Id)
	if err != nil {
		return err
	}
	return nil
}