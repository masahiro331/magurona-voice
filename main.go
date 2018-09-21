package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/user"
	"time"

	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto"
)

var (
	BaseDir = ""
)

const (
	baseurl = "http://magurona-button.sakura.ne.jp/"
	jsonurl = "js/magurona_voice.json"
)

type Category struct {
	Kinds []Kind `json:"category"`
}

type Kind struct {
	Name   string  `json:"name"`
	Voices []Voice `json:"voices"`
}

type Voice struct {
	Title string `json:"title"`
	File  string `json:"file"`
}

func run() error {
	var category Category
	voiceurl := ""
	filedir := ""
	r := -1

	resp, err := http.Get(fmt.Sprintf("%s%s", baseurl, jsonurl))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&category)
	if err != nil {
		return err
	}

	if flag.Arg(0) == "" {
		r = rand.Intn(len(category.Kinds))
	}

	for i, kind := range category.Kinds {
		_ = os.Mkdir(fmt.Sprintf("%s/%s", BaseDir, kind.Name), 0755)
		if i == r || kind.Name == flag.Arg(0) {
			ri := rand.Intn(len(kind.Voices))
			voiceurl = fmt.Sprintf("%s%s", baseurl, kind.Voices[ri].File)
			filedir = fmt.Sprintf("%s/%s/%s.mp3", BaseDir, kind.Name, kind.Voices[ri].Title)
		}
	}

	_, err = os.Stat(filedir)
	if os.IsNotExist(err) {
		file, err := os.Create(filedir)
		if err != nil {
			return err
		}

		v, err := http.Get(voiceurl)
		if err != nil {
			return err
		}
		defer v.Body.Close()
		io.Copy(file, v.Body)
		file.Close()
	} else if err != nil {
		return err
	}

	f, err := os.Open(filedir)
	if err != nil {
		return err
	}
	defer f.Close()

	d, err := mp3.NewDecoder(f)
	if err != nil {
		return err
	}
	defer d.Close()

	p, err := oto.NewPlayer(d.SampleRate(), 2, 2, 8192)
	if err != nil {
		return err
	}
	defer p.Close()

	if _, err := io.Copy(p, d); err != nil {
		return err
	}
	return nil
}

func init() {
	userinfo, err := user.Current()
	if err != nil {
		panic(err)
	}
	BaseDir = fmt.Sprintf("%s/.magurona/voices", userinfo.HomeDir)
	_ = os.MkdirAll(BaseDir, 0755)
}

func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	if err := run(); err != nil {
		log.Fatal(err)
	}
}
