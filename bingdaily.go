package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/caseymrm/menuet"
)

const bingRoot = "https://www.bing.com"
const bingURL = bingRoot + "/HPImageArchive.aspx?format=js&idx=0&n=1"

// TODO: Instead of updating each of these individual variables, pass in a response objecct
var wallpaperTitle = "Updating..."
var wallpaperURL = ""
var quizURL = ""
var description = ""

type bingResponse struct {
	Images []struct {
		URL          string `json:"url"`
		Title        string `json:"title"`
		Copyright    string `json:"copyright"`
		CopyrightURL string `json:"copyrightlink"`
		QuizURL      string `json:"quiz"`
	} `json:"images"`
}

var latestBingResponse bingResponse

func open(url string) {
	exec.Command("open", url).Run()
}

func (br bingResponse) open() {
	exec.Command("open", br.Images[0].URL).Run()
}

func menuItems() []menuet.MenuItem {
	// fmt.Println("called menuItems")
	items := []menuet.MenuItem{
		{
			Text: wallpaperTitle,
			// FontSize: 9,
		},
		// {
		// 	Text: description,
		// },
		{
			Type: menuet.Separator,
		},
	}
	if wallpaperURL != "" {
		items = append(items, menuet.MenuItem{
			Text: "More info...",
			Clicked: func() {
				fmt.Println("Opening", wallpaperURL)
				open(wallpaperURL)
			},
		})
	}
	if quizURL != "" {
		items = append(items, menuet.MenuItem{
			Text: "Quiz link...",
			Clicked: func() {
				open(quizURL)
			},
		})
	}
	return items
}

func getCurrentWallpaperURL() (bingResponse, error) {
	var br bingResponse
	fmt.Println("Downloading new wallpaper data")
	resp, err := http.Get(bingURL)
	if err != nil {
		return br, err
	}

	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)

	err = dec.Decode(&br)
	if err != nil {
		return br, err
	}
	fmt.Println("Wallpaper: ", br)
	return br, nil
}

func downloadBing() {
	filename := fmt.Sprintf("/tmp/bingdaily-%s.jpg", time.Now().Format("2006-01-02"))
	br, err := getCurrentWallpaperURL()
	if err != nil {
		fmt.Println("Could not get wallpaper info: ", err)
		return
	}

	file, err := os.Create(filename)
	if err != nil {
		return
	}
	defer file.Close()

	resp, err := http.Get(bingRoot + br.Images[0].URL)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return
	}

	fmt.Println("Setting wallpaper to: ", filename)
	err = SetFromFile(filename)
	if err != nil {
		fmt.Println("Unable to set desktop :(", err)
		return
	}

	description = br.Images[0].Copyright
	wallpaperTitle = br.Images[0].Title
	wallpaperURL = br.Images[0].CopyrightURL
	quizURL = bingRoot + br.Images[0].QuizURL
}

// SetFromFile uses AppleScript to tell Finder to set the desktop wallpaper to specified file.
func SetFromFile(file string) error {
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
			menuet.App().SetMenuState(&menuet.MenuState{
				Title: "🌄",
			})
			downloadBing()
			menuet.App().MenuChanged()
			time.Sleep(time.Hour)
		}
	}()
	// go func() {
	// 	for {
	// 		menuet.App().SetMenuState(&menuet.MenuState{
	// 			Title: "🌄",
	// 		})
	// 		menuet.App().MenuChanged()
	// 		time.Sleep(time.Second)
	// 	}
	// }()
	// menuet.App().RunApplication()

	app := menuet.App()
	app.Name = "BingDaily"
	app.Label = "com.github.dacort.bingdaily"
	app.Children = menuItems
	// app.AutoUpdate.Version = "v0.1"
	// app.AutoUpdate.Repo = "caseymrm/notafan"
	app.RunApplication()
}
