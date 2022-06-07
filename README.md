[![Go Report Card](https://goreportcard.com/badge/github.com/pescew/goboard)](https://goreportcard.com/report/github.com/pescew/goboard)

# goboard

Cross platform digital signage server

Serves digital signage over http. Accessible through web browsers.

## Images

Image formats supported: .jpg, .jpeg, .png, .gif

Last display date is parsed from filename.

Image naming convention: "YYYY-MM-DD xxxxx.jpg"

## Usage

```
-bg (string): Sets the background color (default "000000")

-border (int): Sets the image border size (default 0)

-border_color (string): Sets the border color (default "FFFFFF")

-dir (string): Sets the image directory (default "img")

-dur (int): Sets the slide duration in seconds (default 30)

-port (int): Sets the web server port (default 8080)

-shuffle (bool): Enables shuffle sort (default false)

-tz (string): Sets the timezone (default "Local")

-update_interval (int): Sets the interval for updating image list in minutes (default 60)
```
