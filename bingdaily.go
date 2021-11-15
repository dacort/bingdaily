package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/caseymrm/menuet"
	"github.com/hako/durafmt"
)

const bingRoot = "https://www.bing.com"
const bingURL = bingRoot + "/HPImageArchive.aspx?format=js&idx=0&n=1"

type bingResponse struct {
	Images []struct {
		URL          string `json:"url"`
		URLBase		 string `json:"urlbase"`
		Title        string `json:"title"`
		Copyright    string `json:"copyright"`
		CopyrightURL string `json:"copyrightlink"`
		QuizURL      string `json:"quiz"`
		Hash         string `json:"hsh"`
	} `json:"images"`
}

type bingWallpaper struct {
	WallpaperTitle string
	ImageURL       string
	SearchURL      string
	QuizURL        string
	Description    string
	Hash           string
	lastCheckedAt  time.Time
}

func (bw bingWallpaper) isDefault() bool {
	return bw.WallpaperTitle == "Updating..."
}

func (bw bingWallpaper) getRelativeupdatedAt() string {
	if bw.isDefault() {
		return "Not yet!"
	}

	timeDiff := time.Since(bw.lastCheckedAt).Round(time.Second)
	return durafmt.Parse(timeDiff).LimitFirstN(2).String() + " ago"
}

func shortDur(d time.Duration) string {
	s := d.String()
	if strings.HasSuffix(s, "m0s") {
		s = s[:len(s)-2]
	}
	if strings.HasSuffix(s, "h0m") {
		s = s[:len(s)-2]
	}
	return s
}

func (bw bingWallpaper) openSearchURL() {
	bw.logAndOpenURL(bw.SearchURL)
}

func (bw bingWallpaper) openQuizURL() {
	bw.logAndOpenURL(bw.QuizURL)
}

func (bw bingWallpaper) logAndOpenURL(url string) {
	log.Println("Opening", latestBingWallpaper.SearchURL)
	exec.Command("open", url).Run()
}

var latestBingWallpaper = bingWallpaper{
	WallpaperTitle: "Updating...",
}

// menuItems will get called every time the menu bar gets clicked on
func menuItems() []menuet.MenuItem {
	items := []menuet.MenuItem{
		{
			Text: latestBingWallpaper.WallpaperTitle,
		},
		{
			Type: menuet.Separator,
		},
	}

	// Only add this data if we've updated the wallpaper
	if !latestBingWallpaper.isDefault() {
		// Image detail and quiz info
		items = append(items, []menuet.MenuItem{
			{
				Text:    "Search on Bing",
				Clicked: latestBingWallpaper.openSearchURL,
			},
			{
				Text:    "Take the quiz!",
				Clicked: latestBingWallpaper.openQuizURL,
			},
			{
				Type: menuet.Separator,
			},
		}...)

		// Image metadata
		items = append(items, []menuet.MenuItem{
			{
				Text: fmt.Sprintf("Last checked %s", latestBingWallpaper.getRelativeupdatedAt()),
			},
			{
				Text:    "Check for new image",
				Clicked: syncWithBing,
			},
			{
				Type: menuet.Separator,
			},
		}...)
	}

	return items
}

func syncWithBing() {
	bwData, err := getLatestWallpaperMetadata()
	if err != nil {
		log.Println("Sorry, there was an error fetching the latest wallpaper metadata...will retry later!", err)
		return
	}

	// No need to update if the hash is the same!
	if bwData.Hash == latestBingWallpaper.Hash {
		log.Println("No update to the Image of the day! Continue on your merry way. :)")
		latestBingWallpaper.lastCheckedAt = time.Now()
		return
	}

	// Try to save the wallpaper to a file. If we don't succeed...we'll just retry in an hour :)
	filename, err := saveWallpaper(bwData)
	if err != nil {
		return
	}

	err = setWallpaperToFile(filename)
	if err != nil {
		return
	}

	// Only shows up when run as an application bundle
	menuet.App().Notification(menuet.Notification{
		Title:        fmt.Sprintf("New Bing Image of the Day"),
		Subtitle:     bwData.WallpaperTitle,
		ActionButton: "Show desktop",
		CloseButton:  "Close",
	})

	latestBingWallpaper = bwData
}

func getLatestWallpaperMetadata() (bw bingWallpaper, err error) {
	var br bingResponse
	log.Println("Downloading new wallpaper data")
	resp, err := http.Get(bingURL)
	if err != nil {
		log.Printf("Error downloading url (%s): %s", bingURL, err)
		return bw, err
	}

	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&br)
	if err != nil {
		return bw, err
	}

	// TODO: Make ... less terrible?
	bw.WallpaperTitle = br.Images[0].Title
	bw.ImageURL = bingRoot + br.Images[0].URLBase + "_UHD.jpg"
	bw.SearchURL = br.Images[0].CopyrightURL
	bw.QuizURL = bingRoot + br.Images[0].QuizURL
	bw.Description = br.Images[0].Copyright
	bw.Hash = br.Images[0].Hash
	bw.lastCheckedAt = time.Now()

	return bw, nil
}

func saveWallpaper(bw bingWallpaper) (filename string, err error) {
	filename = fmt.Sprintf("/tmp/bingdaily-%s.jpg", time.Now().Format("2006-01-02"))

	// Fetch the image
	resp, err := http.Get(bw.ImageURL)
	if err != nil || resp.StatusCode != 200 {
		log.Printf("There was an error downloading %s (%d): %s", bw.ImageURL, resp.StatusCode, err.Error())
		return "", err
	}
	defer resp.Body.Close()

	file, err := os.Create(filename)
	if err != nil {
		log.Printf("Couldn't create file (%s): %s", filename, err.Error())
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		log.Println("Couldn't save the downloaded image to a file:", err)
		return "", err
	}

	return filename, nil
}

// setWallpaperToFile uses AppleScript to tell Finder to set the desktop wallpaper to specified file.
func setWallpaperToFile(file string) error {
	// THis works, but you need to killall Dock
	// sqlite3 ~/Library/Application\ Support/Dock/desktoppicture.db "update data set value = '/Library/Mobile Documents/com~apple~CloudDocs/Wallpaper'" && killall Dock
	cmd := exec.Command("osascript", "-e", `tell application "System Events" to tell every desktop to set picture to `+strconv.Quote(file))
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	fmt.Println("Command output (", strconv.Quote(file), "): ", out.String())
	return err
}

func main() {
	go registerNotification()
	go func() {
		for {
			syncWithBing()
			menuet.App().MenuChanged()
			time.Sleep(time.Hour)
		}
	}()

	app := menuet.App()
	app.SetMenuState(&menuet.MenuState{
		Title: "🌄",
	})
	app.Name = "BingDaily"
	app.Label = "com.github.dacort.bingdaily"
	app.Children = menuItems

	// This needs to be implemented or the process crashes when click on the notification
	app.NotificationResponder = func(id, response string) {
		showDesktop()
	}

	// TODO: For later
	// app.AutoUpdate.Version = "v0.1"
	// app.AutoUpdate.Repo = "caseymrm/notafan"
	app.RunApplication()
}
