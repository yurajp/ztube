package config

import (
	"fmt"
	"regexp"
	"errors"
	"os"
	
  "github.com/yurajp/confy"
)

type Config struct {
	AppDir string
	DirPath string
	Port string
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
	Conf = &Config{wd, dir, prt}
	
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