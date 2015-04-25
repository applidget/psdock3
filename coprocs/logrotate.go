package coprocs

//output are written to a file file_name, when starting, we need to
// 1: watch if file file_name exists.
// 2: if it exists, check for creation date and adjust future rotation + write output to this file
// 3: when we need to rotate, add time stamp to the end of file_name, zip and start logging in a new file
// 5: when creating a new file make sure there is no more the 5 gzipped file

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

var (
	rotationDelay      = 6 * time.Hour //6 hours by default
	archiveRetainCount = 5             //number of log archive to keep
)

type Rotator struct {
	LogFile            string
	RotationDelay      time.Duration
	ArchiveRetainCount int
	ticker             *time.Ticker

	workingDir string
	watching   bool
}

func NewRotator(logFile string) *Rotator {
	dir := filepath.Dir(logFile)
	return &Rotator{LogFile: logFile, workingDir: dir, RotationDelay: rotationDelay, ArchiveRetainCount: archiveRetainCount}
}

// gzipLogFile gzip r current log file by copying it and empty current log file once done.
// it returns the path of the archived file or an error
func (r *Rotator) gzipLogFile() (string, error) {
	//gzip the current log file into a new file
	t := time.Now().Local()
	archive, err := os.Create(fmt.Sprintf("%s_%s.gz", r.LogFile, t.Format("20060102150405")))
	if err != nil {
		return "", err
	}

	defer archive.Close()

	gzipWriter := gzip.NewWriter(archive)
	defer gzipWriter.Close()

	logFile, err := os.Open(r.LogFile)
	if err != nil {
		return "", err
	}

	bufReader := bufio.NewReader(logFile)
	if _, err := bufReader.WriteTo(gzipWriter); err != nil {
		return "", err
	}
	if err := gzipWriter.Flush(); err != nil {
		return "", err
	}

	//empty the current log file
	return archive.Name(), ioutil.WriteFile(r.LogFile, []byte(""), 0600)
}

// keep ArchiveRetainCount on disk and supress others ()
func (r *Rotator) cleanupOldArchives() error {
	archives, err := r.relatedGZFiles()
	if err != nil {
		return err
	}
	// oldest archives are at the start of the array
	for i := len(archives) - 1; i >= r.ArchiveRetainCount; i-- {
		os.Remove(archives[i])
	}
	return nil
}

// relatedGZFiles looks for file that start with r.LogPath and end with .gz in r.LogPath directory
// returns a list of path sorted !!
func (r *Rotator) relatedGZFiles() ([]string, error) {
	files, err := ioutil.ReadDir(r.workingDir)
	if err != nil {
		return nil, err
	}

	relatedFiles := []string{}
	for _, f := range files {
		fName := f.Name()
		if strings.HasPrefix(fName, filepath.Base(r.LogFile)) && filepath.Ext(fName) == ".gz" {
			relatedFiles = append(relatedFiles, path.Join(r.workingDir, fName))
		}
	}
	return relatedFiles, nil
}

func (r *Rotator) StartWatching() {
	r.watching = true
	r.ticker = time.NewTicker(r.RotationDelay)
	for {
		_ = <-r.ticker.C
		if !r.watching {
			return
		}
		log.Info("Log rotation: rotating")
		if _, err := r.gzipLogFile(); err != nil {
			log.Errorf("Log rotation: %v", err)
		}
		if err := r.cleanupOldArchives(); err != nil {
			log.Errorf("Log rotation: %v", err)
		}
	}
}

func (r *Rotator) StopWatching() {
	if r.watching {
		r.ticker.Stop()
	}
	r.watching = false
}
