package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"

	"github.com/ChimeraCoder/anaconda"
	"github.com/e-dard/gev"
)

type Config struct {
	ConsumerKey    string `json:"consumer_key" env:"TWITTER_CONSUMER_KEY"`
	ConsumerSecret string `json:"consumer_secret" env:"TWITTER_CONSUMER_SECRET"`
	AccessToken    string `json:"access_token" env:"TWITTER_ACCESS_TOKEN"`
	AccessSecret   string `json:"access_token_secret" env:"TWITTER_ACCESS_TOKEN_SECRET"`
}

type Archive struct {
	api       *anaconda.TwitterApi
	user      string
	w         io.Writer
	SinceId   int64
	MaxId     int64
	Count     int64
	ExcludeRT bool
	ExcludeAT bool
}

func NewArchive(api *anaconda.TwitterApi, user string, nort, noat bool) *Archive {
	return &Archive{
		w:         os.Stdout,
		api:       api,
		user:      user,
		Count:     200,
		ExcludeRT: nort,
		ExcludeAT: noat,
		SinceId:   -1,
		MaxId:     -1,
	}
}

func (a *Archive) Params() url.Values {
	v := url.Values(make(map[string][]string))
	v.Add("screen_name", a.user)
	v.Add("count", fmt.Sprintf("%d", a.Count))
	v.Add("include_rts", fmt.Sprintf("%v", !a.ExcludeRT))
	v.Add("exclude_replies", fmt.Sprintf("%v", a.ExcludeAT))

	if a.SinceId > -1 {
		v.Add("since_id", fmt.Sprintf("%d", a.SinceId))
	}
	if a.MaxId > -1 {
		v.Add("max_id", fmt.Sprintf("%d", a.MaxId))
	}
	return v
}

// ArchiveAll fetches upto 3200 of the user's most recent tweets.
func (a *Archive) Archive() error {
	results, err := a.api.GetUserTimeline(a.Params())
	if err != nil {
		return err
	}

	// Process results
	for _, tweet := range results {
		if err := json.NewEncoder(a.w).Encode(tweet); err != nil {
			return err
		}
	}

	if len(results) > 0 {
		a.MaxId = results[len(results)-1].Id - 1
		return a.Archive()
	}
	return nil
}

// UpdateSince fetches upto the 3200 most recently published tweets
// published after `since`.
func (a *Archive) Update() error {
	results, err := a.api.GetUserTimeline(a.Params())
	if err != nil {
		return err
	}

	// Process results
	for _, tweet := range results {
		if err := json.NewEncoder(a.w).Encode(tweet); err != nil {
			return err
		}
	}

	if len(results) > 0 {
		a.MaxId = results[len(results)-1].Id - 1
		return a.Update()
	}
	return nil
}

func (a *Archive) UpdateArchive(pth string) error {
	fd, err := os.Open(pth)
	if err != nil {
		return err
	}

	r := bufio.NewReader(fd)
	b, err := r.ReadBytes('\n')
	if err != nil {
		return err
	}

	// Now we can get the since_id from the id of the first tweet in the
	// file.
	var t anaconda.Tweet
	if err := json.Unmarshal(b, &t); err != nil {
		return err
	}

	// Setup a temporary file
	td, err := ioutil.TempFile(os.TempDir(), "")
	if err != nil {
		return err
	}
	// Redirect Archive to temporary file.
	a.w = td
	a.SinceId = t.Id
	// Write latest updates to head of the file
	if err := a.Update(); err != nil {
		return err
	}

	// Copy the line of the file we already read.
	if _, err := td.Write(b); err != nil {
		return err
	}
	// Copy the rest of the existing file into the temporary file.
	if _, err := io.Copy(td, fd); err != nil {
		return err
	}
	if err := fd.Close(); err != nil {
		return err
	}

	// replace the archive file with the temporary one.
	tmppth := td.Name()
	if err := td.Close(); err != nil {
		return err
	}
	mvCmd := exec.Command("mv", tmppth, pth)
	return mvCmd.Run()
}

func main() {
	i := flag.String("c", ".taconfig", "Location of tua JSON configuration.")
	a := flag.String("a", "", "Location of existing user archive.")
	nort := flag.Bool("nort", false, "Exclude retweets.")
	noat := flag.Bool("noat", false, "Exclude mentions.")
	flag.Parse()

	user := flag.Arg(0)
	if user == "" {
		log.Fatal("User argument required.")
	}

	var c Config
	// Attempt to read configuration from file
	fd, err := os.Open(*i)
	if err != nil {
		if *i != ".taconfig" {
			log.Fatal(err)
		}
	} else {
		defer fd.Close()
		if err := json.NewDecoder(fd).Decode(&c); err != nil {
			log.Fatal(err)
		}
	}

	if err := gev.Unmarshal(&c); err != nil {
		log.Fatal(err)
	}

	anaconda.SetConsumerKey(c.ConsumerKey)
	anaconda.SetConsumerSecret(c.ConsumerSecret)
	api := anaconda.NewTwitterApi(c.AccessToken, c.AccessSecret)

	ok, err := api.VerifyCredentials()
	if err != nil {
		log.Fatal(err)
	} else if !ok {
		log.Fatal("Credentials could not be verified.")
	}

	archive := NewArchive(api, user, *nort, *noat)
	if *a != "" {
		// Attempt to open archive and update.
		if err := archive.UpdateArchive(*a); err != nil {
			log.Fatal(err)
		}
		return
	}
	// Run a full archive.
	if err := archive.Archive(); err != nil {
		log.Fatal(err)
	}
}
