package screenshoter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"

	"github.com/rtfmkiesel/drivebyte/internal/logger"
)

type Screenshoter struct {
	browser *rod.Browser
	options *Options
}

type Options struct {
	ChromePath   string // Manually specify a chrome path
	TempPath     string // The temp directory for the browser to use
	Incognito    bool   // Use incognito mode
	Proxy        string // The HTTP proxy string to set
	BrowserSizeH int    // The horizontal size of the screenshot
	BrowserSizeV int    // The vertical size of the screenshot
	UserAgent    string // The user agent to use
	Timeout      int    // The screenshot timeout in seconds
	OutputDir    string // The output dir for the screenshots
	Foreground   bool   // Show the action (non headless mode)
}

func NewScreenshoter(opt *Options) (s *Screenshoter, err error) {
	s = &Screenshoter{
		options: opt,
	}

	if opt.TempPath == "" {
		opt.TempPath, err = os.MkdirTemp(os.TempDir(), "")
		if err != nil {
			return nil, err
		}
	}
	logger.Debug("Using temp dir %s", opt.TempPath)

	var l *launcher.Launcher
	if opt.ChromePath != "" {
		logger.Debug("User wants to use %s", opt.ChromePath)

		if !doesFileExist(opt.ChromePath) {
			return nil, fmt.Errorf("%s not found", opt.ChromePath)
		}
		l = launcher.New().Bin(opt.ChromePath)

	} else {
		logger.Debug("Checking if Chrome is already present in $PATH")
		if path, isPresent := launcher.LookPath(); isPresent {
			logger.Debug("Using %s", path)
			l = launcher.New().Bin(path)

		} else {
			logger.Warning("Chrome not found, using go-rod's Chrome")
			l = launcher.New()
		}
	}

	l = l.
		Headless(!opt.Foreground).
		UserDataDir(opt.TempPath).
		Set("mute-audio").
		Set("no-first-run").
		Set("no-default-browser-check").
		Set("ignore-certificate-errors").
		//Set("disable-gpu").
		Set("disable-sync").
		Set("disable-infobars").
		Set("disable-notifications").
		Set("disable-crash-reporter").
		Set("ignore-certificate-errors")

	if opt.Incognito {
		l = l.Set("incognito")
	}
	if opt.Proxy != "" {
		l = l.Set("proxy-server", opt.Proxy)
	}
	if opt.UserAgent != "" {
		l = l.Set("user-agent", opt.UserAgent)
	}
	if os.Geteuid() == 0 {
		l = l.Set("no-sandbox") // Is required under linux & root
	}

	url, err := l.Launch()
	if err != nil {
		return nil, err
	}
	logger.Debug("Connecting to Chrome at %s", url)

	s.browser = rod.New().ControlURL(url)
	if err := s.browser.Connect(); err != nil {
		return nil, err
	}

	if !doesFileExist(opt.OutputDir) {
		if err := os.MkdirAll(opt.OutputDir, 0750); err != nil {
			return nil, fmt.Errorf("creating folder %s failed: %s", opt.OutputDir, err)
		}
	}

	return s, nil
}

func (s *Screenshoter) TakeScreenshot(url string) error {
	page, err := s.browser.Timeout(time.Duration(s.options.Timeout) * time.Second).Page(proto.TargetCreateTarget{URL: url})
	if err != nil {
		return fmt.Errorf("failed to navigate to %s: %s", url, err)
	}
	logger.Debug("Page %s open", url)

	timeout := page.Timeout(time.Duration(s.options.Timeout) * time.Second)
	if err := timeout.WaitLoad(); err != nil {
		return fmt.Errorf("reached load timeout on %s: %s", url, err)
	}
	if err := timeout.WaitDOMStable(300*time.Millisecond, 0); err != nil {
		return fmt.Errorf("reached DOMStable timeout on %s: %s", url, err)
	}
	logger.Debug("Page %s loaded", url)

	if err := page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Width:  s.options.BrowserSizeH,
		Height: s.options.BrowserSizeV,
	}); err != nil {
		return fmt.Errorf("failed to set viewport on %s: %s", url, err)
	}
	logger.Debug("Set viewport on page %s", url)

	screenshotBytes, err := page.Screenshot(true, &proto.PageCaptureScreenshot{
		Format: proto.PageCaptureScreenshotFormatPng,
	})
	if err != nil {
		return fmt.Errorf("failed to take screenshot of %s: %s", url, err)
	}
	if err = page.Close(); err != nil {
		return fmt.Errorf("failed to close page of %s: %s", url, err)
	}
	logger.Debug("Took screenshot of %s", url)

	// To allow for chaining, output the URL to stdout
	logger.Stdout("%s\n", url)

	if err := s.saveScreenshot(url, screenshotBytes); err != nil {
		return fmt.Errorf("failed to save screenshot of %s: %s", url, err)
	}

	return nil
}

// Cleanup closes the Browser and deletes the temp directory
func (s *Screenshoter) Cleanup() (err error) {
	if err = s.browser.Close(); err != nil {
		return fmt.Errorf("closing browser failed: %s", err)
	}
	logger.Debug("Closed browser")

	// This is needed as browser.Close() takes longer than the return
	time.Sleep(1 * time.Second)

	if err = os.RemoveAll(s.options.TempPath); err != nil {
		return fmt.Errorf("removing temp dir %s failed: %s", s.options.TempPath, err)
	}
	logger.Debug("Removed temp dir %s", s.options.TempPath)

	return nil
}

// saveScreenshot saves a screenshot to the output directory
func (s *Screenshoter) saveScreenshot(url string, screenshotBytes []byte) error {
	filePath := generateFilename(s.options.OutputDir, url)

	file, err := os.Create(filePath) // #nosec G304 as intended functionality
	if err != nil {
		return fmt.Errorf("creating file %s failed: %s", filePath, err)
	}
	defer file.Close()

	if _, err = file.Write(screenshotBytes); err != nil {
		return fmt.Errorf("writing to file %s failed: %s", filePath, err)
	}

	logger.Debug("Saved screenshot of %s to %s", url, filePath)
	return nil
}

// generateFilename will generate the filename from the URL by sanitizing bad chars and trying until a unique filename is found
func generateFilename(dir string, url string) string {
	// Sanitize filename
	unsafeChars := map[rune]bool{
		'/': true, '\\': true, '<': true, '>': true, ':': true,
		'"': true, '|': true, '?': true, '*': true,
	}

	var safeFilename strings.Builder
	underscore := false // To track replacement (no double '_')

	for _, char := range url {
		if !unsafeChars[char] {
			safeFilename.WriteRune(char)
			underscore = false
		} else {
			if !underscore {
				safeFilename.WriteRune('_')
			}
			underscore = true
		}
	}

	filePath := filepath.Join(dir, safeFilename.String()+".png")

	if !doesFileExist(filePath) {
		return filePath
	}

	baseName := safeFilename.String()
	ext := ".png"

	// Iterate until a filename is found that does not exist
	for i := 1; ; i++ {
		newFileName := fmt.Sprintf("%s_%d%s", baseName, i, ext)
		newPath := filepath.Join(dir, newFileName)

		if !doesFileExist(newPath) {
			// Filename does not exist, return new name
			return newPath
		}
	}
}

func doesFileExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}
