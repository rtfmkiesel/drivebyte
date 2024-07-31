package main

import (
	"sync"
	"time"

	"github.com/go-rod/rod/lib/proto"
	"github.com/rtfmkiesel/drivebyte/internal/logger"
	"github.com/rtfmkiesel/drivebyte/internal/options"
	"github.com/rtfmkiesel/drivebyte/internal/webdiscovery"
)

func main() {
	opt, err := options.ParseCliOptions()
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
				page := opt.Browser.
					Timeout(time.Duration(opt.ScreenshotTimeout) * time.Second).
					MustPage(url).
					MustWaitDOMStable()

				page.MustSetViewport(opt.BrowserSizeH, opt.BrowserSizeV, 1, false)

				screenshotBytes, err := page.Screenshot(true, &proto.PageCaptureScreenshot{
					Format: proto.PageCaptureScreenshotFormatPng,
				})
				if err != nil {
					logger.ErrorF("Failed to take screenshot of %s: %s", url, err)
					continue
				}
				if err = page.Close(); err != nil {
					logger.ErrorF("Failed to close page of %s: %s", url, err)
					continue
				}

				// To allow for chaining, output the URL to stdout
				logger.Stdout("%s\n", url)

				if err := opt.SaveScreenshot(url, screenshotBytes); err != nil {
					logger.ErrorF("Failed to save screenshot of %s: %s", url, err)
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

	if err := opt.Cleanup(); err != nil {
		logger.Fatal(err)
	}
}
