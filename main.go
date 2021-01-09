package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"sync"
	"time"

	pdfParse "github.com/rootofevil/lapsnapperpdfparse"

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
	var contentdir string
	var imagefile string

	log.Println("Start, begin to pick up parameters")
	// flag.StringVar(&inputdir, "inputdir", "input", "Dir with source files (pdf)")
	// flag.StringVar(&outdir, "outdir", "out", "Dir for image output")
	// flag.StringVar(&fontsdir, "fontsdir", "fonts", "Dir with fonts")
	// flag.StringVar(&archivedir, "archivedir", "archive", "Dir where to archive files")
	// flag.StringVar(&fontname, "fontname", "arial.ttf", "Font filename")
	// flag.StringVar(&contentdir, "contentdir", "content", "Dir with graphical content")
	// flag.StringVar(&imagefile, "imagefile", "logo.jpg", "Name of image to add")
	flag.StringVar(&access_token, "a", "", "Facebook Page access token")
	flag.StringVar(&pageId, "i", "", "Facebook page ID")
	flag.Parse()

	log.Println("Auth token", access_token)
	log.Println("Pageid", pageId)

	config := loadConf("config.json")
	// access_token = config.Facebook.Token
	// pageId = config.Facebook.PageId

	inputdir = config.Inputdir
	outdir = config.Outdir
	fontsdir = config.Fontsdir
	archivedir = config.Archivedir
	fontname = config.Fontname
	contentdir = config.Contentdir
	imagefile = config.Imagefile

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

	fontfile := path.Join(fontsdir, fontname)
	postimage := path.Join(outdir, "out.jpg")
	picture := path.Join(contentdir, imagefile)

	if _, err := os.Stat(fontfile); os.IsNotExist(err) {
		log.Fatal(err)
	}

	if _, err := os.Stat(picture); os.IsNotExist(err) {
		log.Fatal(err)
	}

	log.Println("Starting to listen files")

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
				err = archiveFile(f.Name(), inputdir, archivedir)
				if err != nil {
					log.Println(err)
				}
				continue
			}
			log.Println("Processing file:", f.Name())
			inputfile := path.Join(inputdir, f.Name())
			session, err := pdfParse.ReadPdf(inputfile)
			if err != nil {
				log.Println(err)
				continue
			}

			// message := sessionToText(session)
			// log.Print(message)
			lines := sessionToLines(session)
			prepareImage(lines, postimage, fontfile, picture)
			wg.Add(1)
			go func() {
				defer wg.Done()
				postId, err := fbPublishPhoto(postimage)
				if err != nil {
					log.Println(err)
				} else {
					log.Println("Published post id:", postId)
				}
			}()
			err = archiveFile(f.Name(), inputdir, archivedir)
			if err != nil {
				log.Println(err)
			}
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

func prepareImage(text []string, out string, fontpath string, imagepath string) {
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
	file, err := os.Open(imagepath)
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

func archiveFile(name, inputdir, archivedir string) (err error) {
	log.Println("Moving to archive:", name)
	sourcePath := path.Join(inputdir, name)
	destPath := path.Join(archivedir, name)
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("Couldn't open source file: %s", err)
	}
	outputFile, err := os.Create(destPath)
	if err != nil {
		inputFile.Close()
		return fmt.Errorf("Couldn't open dest file: %s", err)
	}
	defer outputFile.Close()
	_, err = io.Copy(outputFile, inputFile)
	inputFile.Close()
	if err != nil {
		return fmt.Errorf("Writing to output file failed: %s", err)
	}
	// The copy was successful, so now delete the original file
	err = os.Remove(sourcePath)
	if err != nil {
		return fmt.Errorf("Failed removing original file: %s", err)
	}
	return nil
}
