package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	//portFlag is used to set the web server port
	portFlag = flag.Int("port", 8080, "Sets the web server port")

	//directoryFlag is used to set the web server port
	directoryFlag = flag.String("dir", "img", "Sets the image directory")

	//shuffleFlag is used to set the web server port
	shuffleFlag = flag.Bool("shuffle", false, "Enables shuffle sort (default false)")

	//durationFlag is used to set the slide duration in seconds
	durationFlag = flag.Int("dur", 30, "Sets the slide duration in seconds")

	//tzFlag is used to set the time zone
	tzFlag = flag.String("tz", "Local", "Sets the timezone")

	//bgFlag is used to set the background color
	bgFlag = flag.String("bg", "000000", "Sets the background color")

	//borderFlag is used to set the the image border size
	borderFlag = flag.Int("border", 0, "Sets the image border size")

	//borderColorFlag is used to set the border color
	borderColorFlag = flag.String("border_color", "255, 255, 255", "Sets the border color")

	//intervalFlag is used to set the update interval in minutes
	intervalFlag = flag.Int("update_interval", 60, "Sets the interval for updating image list in minutes")

	//strings
	borderSwitch, borderSize, durationString, divContent, intervalString string

	//timezone
	localTZ *time.Location

	//errors
	err error
)

func main() {
	rand.Seed(time.Now().UnixNano())

	//parse args
	flag.Parse()
	dirPath := "./" + *directoryFlag
	fmt.Println("Using image directory:", dirPath)
	if *shuffleFlag {
		fmt.Println("Shuffle enabled")
	} else {
		fmt.Println("Shuffle disabled")
	}

	//slide duration check
	if *durationFlag < 0 || *durationFlag > 86400 {
		fmt.Println("ERROR: Slide duration must be between 0 and 86400 secconds (24 hours)")
		fmt.Println("DURATION SET TO 30 SECONDS")
		durationString = strconv.Itoa(30 * 1000)
	} else {
		fmt.Println("Slide duration set to", *durationFlag, "seconds")
		durationString = strconv.Itoa(*durationFlag * 1000)
	}

	//update interval check
	if *intervalFlag < 1 || *intervalFlag > 43800 {
		fmt.Println("ERROR: Update interval must be between 1 and 43800 minutes (1 month)")
		fmt.Println("UPDATE INTERVAL SET TO 1 HOUR")
		*intervalFlag = 60
		intervalString = strconv.Itoa(3600)
	} else {
		fmt.Println("Update interval set to", *intervalFlag, "minutes")
		intervalString = strconv.Itoa(*intervalFlag * 60)
	}

	//image border check
	if *borderFlag == 0 {
		borderSwitch = "0"
		borderSize = "0"
	} else if *borderFlag < 0 || *borderFlag > 300 {
		fmt.Println("ERROR: Border must be between 0 and 300")
		fmt.Println("BORDER DISABLED")
		borderSwitch = "0"
		borderSize = "0"
	} else {
		borderSwitch = "1"
		borderSize = strconv.Itoa(*borderFlag)
	}

	//timezone check
	if *tzFlag != "Local" {
		fmt.Println("Using manual timezone:", *tzFlag)
	}
	localTZ, err = time.LoadLocation(*tzFlag)
	if err != nil {
		log.Fatal(err)
	}

	//http port check
	if *portFlag < 1 || *portFlag > 65535 {
		fmt.Println("ERROR: Port number must be between 1 and 65535")
		fmt.Println("USING DEFAULT PORT 8080")
		*portFlag = 8080
	}
	fmt.Println("Server running on HTTP port", *portFlag)
	portString := ":" + strconv.Itoa(*portFlag)
	fmt.Printf("http://localhost%s\n", portString)

	//populate image list
	fmt.Println("---------------------")
	updateImages()

	//start http server
	http.HandleFunc("/", mainServer)
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir(dirPath))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./js"))))
	go http.ListenAndServe(portString, nil)

	//update ticker
	ticker := time.NewTicker(time.Duration(*intervalFlag) * time.Minute)
	for {
		select {
		case t := <-ticker.C:
			fmt.Println("---------------------")
			fmt.Println("Image list updated", t)
			updateImages()
		}
	}
}

//http server
func mainServer(w http.ResponseWriter, r *http.Request) {
	htmlString := fmt.Sprintf(htmlPage, intervalString, borderSize, *borderColorFlag, borderSwitch, *bgFlag, divContent, durationString)
	fmt.Fprint(w, htmlString)
}

//reads files from directory
func readDirectory() []os.FileInfo {
	files, err := ioutil.ReadDir(*directoryFlag)
	if err != nil {
		log.Fatal(err)
	}

	return files
}

//shuffles string slice
func shuffleString(list []string) []string {
	rand.Shuffle(len(list), func(i, j int) { list[i], list[j] = list[j], list[i] })
	return list
}

//lists all files in directory matching given extension
func walkMatch(root, pattern string) ([]string, []string, error) {
	var matches []string
	var matchesRaw []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return err
		} else if matched {
			matches = append(matches, path)
			matchesRaw = append(matchesRaw, filepath.Base(path))
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return matches, matchesRaw, nil
}

func updateImages() {
	//calculate elapsed time today
	tdHr, tdMin, tdSec := time.Now().Clock()
	tdLapsed := int64((tdHr * 3600) + (tdMin * 60) + tdSec)

	//load images of various formats
	jpgListFull, jpgList, err := walkMatch(*directoryFlag, "*.jpg")
	if err != nil {
		log.Fatal(err)
	}
	jpegListFull, jpegList, err := walkMatch(*directoryFlag, "*.jpeg")
	if err != nil {
		log.Fatal(err)
	}
	pngListFull, pngList, err := walkMatch(*directoryFlag, "*.png")
	if err != nil {
		log.Fatal(err)
	}
	gifListFull, gifList, err := walkMatch(*directoryFlag, "*.gif")
	if err != nil {
		log.Fatal(err)
	}
	_, _, _, _ = jpgListFull, jpegListFull, pngListFull, gifListFull

	//combine slices
	var imgList []string
	imgList = append(imgList, jpgList...)
	imgList = append(imgList, jpegList...)
	imgList = append(imgList, pngList...)
	imgList = append(imgList, gifList...)

	//shuffle or sort alphabetically
	if *shuffleFlag {
		shuffleString(imgList)
	} else {
		sort.Strings(imgList)
	}

	//extract dates from filenames
	var dateStringList, imgListFinal []string
	for i := 0; i < len(imgList); i++ {
		dateStringList = append(dateStringList, strings.Fields(imgList[i])[0])
		parsedDate, err := time.ParseInLocation("2006-01-02", dateStringList[i], localTZ)
		if err != nil {
			fmt.Printf("Could not parse date from filename \"%s\"...skipping\n", imgList[i])
		}
		if parsedDate.Unix() >= time.Now().Unix()-tdLapsed {
			imgListFinal = append(imgListFinal, imgList[i])
		}
	}

	//generate html div elements
	divContent = ""
	for i := 0; i < len(imgListFinal); i++ {
		fmt.Println(imgListFinal[i])
		divContent += "<div><img src=\"img/" + imgListFinal[i] + "\"></div>\n"
	}
}

//html content
const htmlPage = `
<html>
<head>
<meta http-equiv="refresh" content="%v" />
<style>
#slideshow {
display: block;
height: 98vh;
margin-top 0;
margin-right: auto;
margin-bottom: 0;
margin-left: auto;
max-width:98vw;
max-height:98vh;
}
#slideshow img {
display: block;
height: 98vh;
margin-top 0;
margin-right: auto;
margin-bottom: 0;
margin-left: auto;
box-shadow: 0 0 %vpx rgba(%v,%v); 
}
body {
overflow:hidden;
}
</style>
</head>
<body bgcolor="%v" style="cursor: none">
<script src="js/jquery-3.4.1.min.js"></script>
<div id="slideshow">
   %v
</div>
<script type="text/javascript">
$("#slideshow > div:gt(0)").hide();
setInterval(function() { 
  $('#slideshow > div:first')
    .fadeOut(0)
    .next()
    .fadeIn(1000)
    .end()
    .appendTo('#slideshow');
},  %v);
</script>
</body>
</html>
`
