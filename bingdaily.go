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
	"time"

	"github.com/caseymrm/menuet"
)

const bingRoot = "https://www.bing.com"
const bingURL = bingRoot + "/HPImageArchive.aspx?format=js&idx=0&n=1"

type bingResponse struct {
	Images []struct {
		URL          string `json:"url"`
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
	Descriptiong   string
	Hash           string
}

var latestBingWallpaper = bingWallpaper{
	WallpaperTitle: "Updating...",
}

// menuItems will get called every time the menu bar gets clicked on
func menuItems() []menuet.MenuItem {
	fmt.Println("called menuItems")
	items := []menuet.MenuItem{
		{
			Text: latestBingWallpaper.WallpaperTitle,
		},
		{
			Type: menuet.Separator,
		},
	}
	if latestBingWallpaper.SearchURL != "" {
		items = append(items, menuet.MenuItem{
			Text: "More info...",
			Clicked: func() {
				fmt.Println("Opening", latestBingWallpaper.SearchURL)
				open(latestBingWallpaper.SearchURL)
			},
		})
	}
	if latestBingWallpaper.QuizURL != "" {
		items = append(items, menuet.MenuItem{
			Text: "Quiz link...",
			Clicked: func() {
				open(latestBingWallpaper.QuizURL)
			},
		})
	}
	return items
}

func open(url string) {
	exec.Command("open", url).Run()
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
	bw.ImageURL = bingRoot + br.Images[0].URL
	bw.SearchURL = br.Images[0].CopyrightURL
	bw.QuizURL = bingRoot + br.Images[0].QuizURL
	bw.Descriptiong = br.Images[0].Copyright
	bw.Hash = br.Images[0].Hash

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
	go func() {
		for {
			syncWithBing()
			menuet.App().MenuChanged()
			time.Sleep(time.Second)
		}
	}()

	app := menuet.App()
	app.SetMenuState(&menuet.MenuState{
		Title: "🌄",
	})
	app.Name = "BingDaily"
	app.Label = "com.github.dacort.bingdaily"
	app.Children = menuItems

	// TODO: For later
	// app.AutoUpdate.Version = "v0.1"
	// app.AutoUpdate.Repo = "caseymrm/notafan"
	app.RunApplication()
}
