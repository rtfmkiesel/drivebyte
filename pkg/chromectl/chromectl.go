// chromectl is a package for creating screenshots with the chrome browser
package chromectl

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/rtfmkiesel/drivebyte/pkg/logger"
	"github.com/rtfmkiesel/drivebyte/pkg/options"
)

// Browser is a handle to a webbrowser, ready to take screenshots.
type Browser struct {
	Opt *options.Options
}

// Screenshot() makes a screenshot of the website provided via the "url" parameter.
// A full URL like 'https://example.com' or 'https://sub.example.com:8443' is required.
func (d *Browser) Screenshot(url string) (imgPath string, err error) {
	imgName := createFilename(url) + ".png" // Remove bad characters from the URL to get a valid filename
	imgPath = filepath.Join(d.Opt.OutputDir, imgName)
	logger.Info("Trying to screenshot '%s' to '%s'", url, imgPath)

	args := []string{
		"--headless",
		"--hide-scrollbars",
		"--mute-audio",
		"--no-first-run",
		"--no-default-browser-check",
		"--ignore-certificate-errors", // Always -> since many sites can have invalid SSL from leftovers
		"--disable-gpu",
		"--disable-sync",
		"--disable-infobars",
		"--disable-notifications",
		"--disable-crash-reporter",
		"--user-data-dir=" + d.Opt.TempDir,
		"--window-size=" + fmt.Sprintf("%d,%d", d.Opt.SizeH, d.Opt.SizeV),
		"--screenshot=" + imgPath,
	}

	// Set options if selected
	if d.Opt.IncognitoMode {
		args = append(args, "--incognito")
	}
	if d.Opt.Proxy != "" {
		args = append(args, fmt.Sprintf("--proxy-server=%s", d.Opt.Proxy))
	}
	if d.Opt.UserAgent != "" {
		args = append(args, fmt.Sprintf("--user-agent=%s", d.Opt.UserAgent))
	} else {
		args = append(args, fmt.Sprintf("--user-agent=%s", randomUserAgent()))
	}
	if os.Geteuid() == 0 {
		args = append(args, "--no-sandbox") // Is required under linux & root
	}

	args = append(args, url) // URL must be the last argument

	// Generate context to be able to kill process after a specific timeout
	timeout := time.Duration(d.Opt.ScreenshotTimeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, d.Opt.ChromePath, args...)
	err = cmd.Start()
	if err != nil {
		return "", fmt.Errorf("screenshot failed: '%s'", err)
	}

	// Create a background task that fetches the signal from cmd.Wait() so it can be used in the select statement below
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		err = cmd.Process.Kill()
		if err != nil {
			return "", fmt.Errorf("screenshot timed out, could not kill process: '%s'", err)
		}
		return "", fmt.Errorf("screenshot timed out")
	case err := <-done:
		if err != nil {
			return "", fmt.Errorf("screenshot finished but got error: '%s'", err)
		}
	}

	// Chrome saves all screenshots as 0600, be more permissive here
	err = os.Chmod(imgPath, 0644)
	if err != nil {
		// chmod failed
		return "", fmt.Errorf("setting permissions on screenshot failed: '%s'", err)
	}

	logger.Success("Screenshot of '%s' successful ", url)
	return imgPath, nil
}

// Clean() will delete the temp data generated by the browser.
func (d *Browser) Clean() (err error) {
	err = os.RemoveAll(d.Opt.TempDir)
	if err != nil {
		return fmt.Errorf("deleting temp dir '%s' failed: '%s'", d.Opt.TempDir, err)
	}

	logger.Info("Removed temp dir '%s'", d.Opt.TempDir)
	return nil
}

// randomUserAgent() returns a random user agent from a predefined list taken from useragent.me/api.
func randomUserAgent() string {
	agents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36 Edg/110.0.1587.63",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/110.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36 Edg/110.0.1587.57",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.3 Safari/605.1.15",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36 OPR/95.0.0.0",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36",
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/110.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36 Edg/110.0.1587.41",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36 Edg/110.0.1587.56",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.5060.53 Safari/537.36 Edg/103.0.1264.37",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:102.0) Gecko/20100101 Firefox/102.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.85 Safari/537.36 Edg/90.0.818.46",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36 Edg/110.0.1587.50",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Whale/3.19.166.16 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/110.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36 Edg/108.0.1462.76",
		"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.71 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36 Edg/110.0.1587.46",
		"Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:109.0) Gecko/20100101 Firefox/110.0",
		"Mozilla/5.0 (Windows NT 6.1; Trident/7.0; rv:11.0) like Gecko",
		"Mozilla/5.0 (Windows NT 6.3; Win64; x64; rv:109.0) Gecko/20100101 Firefox/110.0",
		"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.71 Safari/537.36 Core/1.94.192.400 QQBrowser/11.5.5250.400",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36 Edg/109.0.1518.78",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3865.120 Safari/537.36",
		"Mozilla/5.0 (Windows NT 6.3; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36 OPR/95.0.0.0",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36 Edg/110.0.1587.63",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36 Edg/92.0.902.67",
		"Mozilla/5.0 (Windows NT 6.1; WOW64; Trident/7.0; rv:11.0) like Gecko",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:108.0) Gecko/20100101 Firefox/108.0",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36",
		"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.141 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/64.0.3282.140 Safari/537.36 Edge/18.17763",
		"Mozilla/5.0 (X11; Linux x86_64; rv:108.0) Gecko/20100101 Firefox/108.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36 Edg/110.0.1587.63",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/111.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36 Edg/108.0.1462.54",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36 Edg/109.0.1518.61",
		"Mozilla/5.0 (Windows NT 10.0; rv:109.0) Gecko/20100101 Firefox/110.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36 Edg/109.0.1518.70",
	}

	randInt := rand.Intn(len(agents))
	return agents[randInt]
}

// createFilename() will remove all chars (runes) from a string that it becomes a valid file name
func createFilename(s string) string {
	m := make(map[rune]bool)
	for _, char := range []rune{'/', '\\', '<', '>', ':', '"', '|', '?', '*'} {
		m[char] = true
	}

	new := []rune{}
	for _, char := range s {
		if !m[char] {
			new = append(new, char)
		} else {
			new = append(new, '_')
		}
	}

	return string(new)
}
