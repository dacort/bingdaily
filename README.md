# BingDaily

A fun little side project that downloads the Bing Image of the day and makes it your background on macOS.

- Inspiration: Hammerspoon's [BingDaily](https://www.hammerspoon.org/Spoons/BingDaily.html) spoon.
- Setting macOS Desktop Image: https://www.tech-otaku.com/mac/setting-desktop-image-macos-mojave-from-command-line/

## Running

For now, either build with `go build && ./bingdaily` or run the app directly with `go run bingdaily.go`.

On first run, it will download today's Bing image and make it your background.

You can see the name of the image in your new menu bar item and get more info or open today's quiz using the appropriate links.

It will attempt to update the background every hour while the app is running.

## ToDos

- ~~Add "last updated" menu item~~
- ~~Add "check for update" menu item~~
