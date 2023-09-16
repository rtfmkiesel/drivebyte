package main

import (
	"os"
	"sync"

	"drivebyte/pkg/logger"
	"drivebyte/pkg/options"
	"drivebyte/pkg/runner"
	"drivebyte/pkg/webdiscovery"
)

func main() {
	opt, err := options.Parse()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	chanUrls := make(chan string)
	wgUrls := new(sync.WaitGroup)
	chanScreenshots := make(chan string)
	wgScreenshots := new(sync.WaitGroup)

	for i := 0; i < opt.Workers; i++ {
		go runner.Urls(wgUrls, chanUrls, chanScreenshots, opt)
		wgUrls.Add(1)
	}

	for i := 0; i < opt.Workers; i++ {
		go runner.Screenshots(wgScreenshots, chanScreenshots, opt)
		wgScreenshots.Add(1)
	}

	for _, url := range webdiscovery.Generate(opt) {
		chanUrls <- url
	}

	close(chanUrls)
	wgUrls.Wait()
	close(chanScreenshots)
	wgScreenshots.Wait()
}
