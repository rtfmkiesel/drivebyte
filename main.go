package main

import (
	"sync"

	"github.com/rtfmkiesel/drivebyte/internal/logger"
	"github.com/rtfmkiesel/drivebyte/internal/options"
	"github.com/rtfmkiesel/drivebyte/internal/webdiscovery"
)

func main() {
	opt, screenshoter, err := options.Parse()
	if err != nil {
		logger.Fatal(err)
	}

	urls := make(chan string)
	wgWebdiscovery := new(sync.WaitGroup)
	toScreenshot := make(chan string)
	wgChrome := new(sync.WaitGroup)

	for i := 0; i < opt.Workers; i++ {
		wgWebdiscovery.Add(1)
		go func() {
			defer wgWebdiscovery.Done()

			for url := range urls {
				if webdiscovery.IsReachable(url, opt.PortTimeout) {
					toScreenshot <- url
				}
			}
		}()
	}

	for i := 0; i < opt.Workers; i++ {
		wgChrome.Add(1)
		go func() {
			defer wgChrome.Done()

			for url := range toScreenshot {
				if err := screenshoter.TakeScreenshot(url); err != nil {
					logger.Error(err)
					continue
				}
			}
		}()
	}

	for _, url := range webdiscovery.GenerateUrls(opt.Domains, opt.Ports) {
		urls <- url
	}

	close(urls)
	wgWebdiscovery.Wait()
	close(toScreenshot)
	wgChrome.Wait()

	if err := screenshoter.Cleanup(); err != nil {
		logger.Fatal(err)
	}
}
