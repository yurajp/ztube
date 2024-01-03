package config

import (
	"fmt"
	"log"
	"regexp"
	"errors"
	"os"
	
  "github.com/yurajp/confy"
)

func init() {
	err := Prepare()
	if err != nil {
		log.Fatal(err)
	}
}


type Config struct {
	AppDir string
	DirPath string
	Port string
	ShareDir string
	ImgDir string
}

var Conf *Config

func SetConfPath() {
	wd, _ := os.Getwd()
	cnf := "/config/config.ini"
	confy.SetPath(wd + cnf)
}

func SetConfTerm() error {
	wd, _ := os.Getwd()
	SetConfPath()
	fmt.Println("  Type path to directory\n  where files will be saved\n  ")
	var dir string
	fmt.Scanf("%s", &dir)

	todo := true
  var prt string
	for todo {
	  fmt.Println("  Type port number\n  or enter for default\n  (3663)\n  ")
	  fmt.Scanf("%s", &prt)
	  if prt == "" {
	  	prt = "3663"
	  }
	  re := regexp.MustCompile(`[0-9]{4,5}`)
	  if re.MatchString(prt) {
	  	todo = false
	  }
	}
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0750)
		if err != nil {
			return err
		}
		fmt.Printf("  Directory '%s' was created\n", dir)
	}
	
	var shd string
	fmt.Println("  Type path to folder for sharing audio (optionally)\n  ")
	fmt.Scanf("%s", &shd)
	
	var imd string
	fmt.Println("  Type path to folder for store images (optionally)\n  ")
	fmt.Scanf("%s", &imd)
	
	Conf = &Config{wd, dir, prt, shd, imd}
	
	err = confy.WriteConfy(*Conf)
	if err != nil {
		return err
	}
	fmt.Println("  Config saved")
	return nil
}

func LoadConfig() error {
	SetConfPath()
	iface, err := confy.LoadConfy(Config{})
	if err != nil {
		return err
	}
	c, ok := iface.(Config)
	if !ok {
		return errors.New("Cannot convert to Config")
	}
	Conf = &c
	return nil
}

func Prepare() error {
	SetConfPath()
	if !confy.ConfigExists() {
		err := SetConfTerm()
		if err != nil {
			return fmt.Errorf("WriteConfError: %s", err)
		}
	}
	err := LoadConfig()
	if err != nil {
		return fmt.Errorf("LoadConfError: %s", err)
	}
	return nil
}