package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"pdfParse"
	"regexp"
	"sync"
	"time"

	"github.com/golang/freetype"
	"github.com/hqbobo/text2pic"
)

var access_token string
var pageId string

func main() {
	var wg sync.WaitGroup

	var inputdir string
	var outdir string
	var fontsdir string
	var archivedir string
	var fontname string

	flag.StringVar(&inputdir, "inputdir", "input", "Dir with source files (pdf)")
	flag.StringVar(&outdir, "outdir", "out", "Dir for image output")
	flag.StringVar(&fontsdir, "fontsdir", "fonts", "Dir with fonts")
	flag.StringVar(&archivedir, "archivedir", "archive", "Dir where to archive files")
	flag.StringVar(&fontname, "fontname", "arial.ttf", "Font filename")
	flag.Parse()

	if _, err := os.Stat(outdir); os.IsNotExist(err) {
		err := os.Mkdir(outdir, 0700)
		if err != nil {
			log.Println(err)
		}
	}

	if _, err := os.Stat(archivedir); os.IsNotExist(err) {
		err := os.Mkdir(archivedir, 0700)
		if err != nil {
			log.Println(err)
		}
	}

	if _, err := os.Stat(inputdir); os.IsNotExist(err) {
		err := os.Mkdir(inputdir, 0700)
		if err != nil {
			log.Println(err)
		}
	}

	fontfile := path.Join(fontsdir, "arial.ttf")
	postimage := path.Join(outdir, "out.jpg")

	if _, err := os.Stat(fontfile); os.IsNotExist(err) {
		log.Fatal(err)
	}

	config := loadConf("config.json")
	access_token = config.Facebook.Token
	pageId = config.Facebook.PageId

	for {
		files, err := ioutil.ReadDir(inputdir)
		if err != nil {
			log.Println(err)
			continue
		}
		r, _ := regexp.Compile(`.+\.pdf`)

		for _, f := range files {
			if !r.MatchString(f.Name()) {
				log.Println("Wrong file:", f.Name())
				archiveFile(f.Name(), inputdir, archivedir)
				continue
			}
			log.Println("Processing file:", f.Name())
			inputfile := path.Join(inputdir, f.Name())
			session, err := pdfParse.ReadPdf(inputfile)
			if err != nil {
				log.Println(err)
			}

			// message := sessionToText(session)
			// log.Print(message)
			lines := sessionToLines(session)
			prepareImage(lines, postimage, fontfile)
			wg.Add(1)
			go func() {
				defer wg.Done()
				postId, err := fbPublishPhoto(postimage)
				if err != nil {
					log.Println(err)
				}
				log.Println("Published post id:", postId)
			}()
			archiveFile(f.Name(), inputdir, archivedir)
		}
		wg.Wait()
		time.Sleep(5000 * time.Millisecond)
	}

}

func sessionToText(session pdfParse.RaceSession) string {
	text := fmt.Sprintf("%v\nStarted: %v\nEnded: %v\n\nPosition\tCar\t\tBest time\tDif\tTotal time\tLaps\n", session.Type, session.Started, session.Ended)
	for _, r := range session.TimeAttackResults {
		line := fmt.Sprintf("%v\t\t%v\t%v\t%v\t%v\t%v\n", r.Position, r.Driver, r.BestTime, r.Dif, r.TotalTime, r.Laps)
		text += line
	}
	return text
}

func sessionToLines(session pdfParse.RaceSession) []string {
	var lines []string
	lines = append(lines, fmt.Sprintf("%v", session.Type))
	lines = append(lines, fmt.Sprintf("Started: %v", session.Started))
	lines = append(lines, fmt.Sprintf("Ended:  %v", session.Ended))
	lines = append(lines, " ")
	lines = append(lines, "Pos. Car             Best time     Dif         Total time   Laps")
	for _, r := range session.TimeAttackResults {
		line := fmt.Sprintf("%v.     %v   %v   %v   %v   %v", r.Position, r.Driver, r.BestTime, r.Dif, r.TotalTime, r.Laps)
		lines = append(lines, line)
	}
	return lines
}

func prepareImage(text []string, out string, fontpath string) {
	// Read the font data.
	fontBytes, err := ioutil.ReadFile(fontpath)
	if err != nil {
		log.Println(err)
		return
	}
	//produce the fonttype
	f, err := freetype.ParseFont(fontBytes)

	if err != nil {
		log.Println(err)
		return
	}
	pic := text2pic.NewTextPicture(text2pic.Configure{Width: 720, BgColor: text2pic.ColorBlack})
	pic.AddTextLine(" ", 8, f, text2pic.ColorBlack, text2pic.Padding{})
	for _, l := range text {
		pic.AddTextLine(l, 6, f, text2pic.ColorWhite, text2pic.Padding{
			Left:      40,
			Right:     20,
			Bottom:    0,
			Top:       0,
			LineSpace: 0,
		})
	}
	pic.AddTextLine(" ", 6, f, text2pic.ColorBlack, text2pic.Padding{})
	file, err := os.Open("content/logo.jpg")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	pic.AddPictureLine(file, text2pic.Padding{Bottom: 40})

	outFile, err := os.Create(out)
	if err != nil {
		return
	}
	defer outFile.Close()
	b := bufio.NewWriter(outFile)
	//produce the output
	err = pic.Draw(b, text2pic.TypeJpeg)
	if err != nil {
		log.Print(err.Error())
	}
	e := b.Flush()
	if e != nil {
		fmt.Println(e)
	}
}

func archiveFile(name, inputdir, archivedir string) {
	err := os.Rename(path.Join(inputdir, name), path.Join(archivedir, name))
	if err != nil {
		log.Println(err)
	}
	log.Println("Moved to archive:", name)
}
