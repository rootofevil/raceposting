package main

import (
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
	footerfilename := path.Join(contentdir, imagefile)

	if _, err := os.Stat(fontfile); os.IsNotExist(err) {
		log.Fatal(err)
	}

	if _, err := os.Stat(footerfilename); os.IsNotExist(err) {
		log.Fatal(err)
	}

	log.Println("Starting to listen files")

	for {
		currentTime := time.Now()
		currentdate := currentTime.Format("2006-01-02")
		currentdir := path.Join(inputdir, currentdate)
		if _, err := os.Stat(currentdir); os.IsNotExist(err) {
			time.Sleep(5000 * time.Millisecond)
			continue
		}
		files, err := ioutil.ReadDir(currentdir)
		if err != nil {
			log.Println(err)
			time.Sleep(5000 * time.Millisecond)
			continue
		}
		r, _ := regexp.Compile(`.+\.pdf`)

		for _, f := range files {
			if !r.MatchString(f.Name()) {
				log.Println("Wrong file:", f.Name())
				err = archiveFile(f.Name(), currentdir, archivedir)
				if err != nil {
					log.Println(err)
				}
				time.Sleep(5000 * time.Millisecond)
				continue
			}
			log.Println("Processing file:", f.Name())
			inputfile := path.Join(currentdir, f.Name())

			params := pdfParse.Parameters{
				InputfilePath:   inputfile,
				FontfilePath:    fontfile,
				FooterImagePath: footerfilename,
				OutputimagePath: postimage,
				FontSize:        5.5,
			}
			rs, err := pdfParse.NewRaceSession(params)
			if err != nil {
				log.Println(err)
				time.Sleep(5000 * time.Millisecond)
				continue
			}

			err = rs.PdfToImage()
			if err != nil {
				log.Println(err)
				time.Sleep(5000 * time.Millisecond)
				continue
			}
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
			err = archiveFile(f.Name(), currentdir, archivedir)
			if err != nil {
				log.Println(err)
			}
		}
		wg.Wait()
		time.Sleep(5000 * time.Millisecond)
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
