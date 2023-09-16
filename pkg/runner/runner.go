// runner handles the goroutines
package runner

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"sync"

	"drivebyte/pkg/chromectl"
	"drivebyte/pkg/logger"
	"drivebyte/pkg/options"
	"drivebyte/pkg/webdiscovery"
)

// Urls() is the goroutine for scanning a domain for open web ports and returning URLs to screenshot.
func Urls(wg *sync.WaitGroup, chanIn <-chan string, chanOut chan<- string, opt options.Options) {
	defer wg.Done()

	for job := range chanIn {
		if webdiscovery.IsOpen(job, opt.PortTimeout) {
			chanOut <- job
		}
	}
}

// Screenshots() is the goroutine for making screenshots of URLs.
func Screenshots(wg *sync.WaitGroup, chanIn <-chan string, opt options.Options) {
	defer wg.Done()
	var err error

	// Set up the browser controller
	chrome := chromectl.Browser{
		Opt: &opt,
	}

	// Set a temp directory
	if chrome.Opt.TempDir == "" {
		chrome.Opt.TempDir = fmt.Sprintf("./temp_%s", randString(8))
		chrome.Opt.TempDir, err = filepath.Abs(chrome.Opt.TempDir)
		if err != nil {
			logger.Error("getting abosulte path of '%s' failed: '%s'", chrome.Opt.TempDir, err)
			return
		}
	} else {
		chrome.Opt.TempDir = filepath.Join(chrome.Opt.TempDir, randString(8))
		chrome.Opt.TempDir, err = filepath.Abs(chrome.Opt.TempDir)
		if err != nil {
			logger.Error("getting abosulte path of '%s' failed: '%s'", chrome.Opt.TempDir, err)
			return
		}
	}
	err = options.CreateDirIfNotExist(chrome.Opt.TempDir)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	for job := range chanIn {
		_, err := chrome.Screenshot(job)
		if err != nil {
			logger.Error(err.Error())
			continue
		} else if !logger.Verbose { // Only print if not verbose
			fmt.Println(job)
		}
	}

	err = chrome.Clean()
	if err != nil {
		logger.Error(err.Error())
		return
	}
}

// randString() will return a "random" string with a fixed length
func randString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}
