package options

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/projectdiscovery/goflags"
	"github.com/rtfmkiesel/drivebyte/internal/logger"
)

var (
	// from https://github.com/michenriksen/aquatone/blob/master/core/ports.go
	portsMin     = []int{80, 443}
	portsDefault = []int{80, 443, 8000, 8080, 8443}
	portsLarge   = []int{80, 81, 443, 591, 2082, 2087, 2095, 2096, 3000, 8000, 8001,
		8008, 8080, 8083, 8443, 8834, 8888}
	portsMax = []int{80, 81, 300, 443, 591, 593, 832, 981, 1010, 1311,
		2082, 2087, 2095, 2096, 2480, 3000, 3128, 3333, 4243, 4567,
		4711, 4712, 4993, 5000, 5104, 5108, 5800, 6543, 7000, 7396,
		7474, 8000, 8001, 8008, 8014, 8042, 8069, 8080, 8081, 8088,
		8090, 8091, 8118, 8123, 8172, 8222, 8243, 8280, 8281, 8333,
		8443, 8500, 8834, 8880, 8888, 8983, 9000, 9043, 9060, 9080,
		9090, 9091, 9200, 9443, 9800, 9981, 12443, 16080, 18091, 18092,
		20720, 28017}
)

// Commandd line options
type Options struct {
	Domains              []string     // The domains to test for screenshots
	Ports                []int        // The ports to test for potential web servers
	PortTimeout          int          // The timeout for port checks
	OutputDir            string       // The output directory for the screenshots TODO
	Workers              int          // The number of concurrent workers
	Browser              *rod.Browser // The rod browser used to do the screenshots
	ScreenshotTimeout    int          // The timeout for the screenshot
	browserPath          string       // The path to the chrome binary
	browserIncognitoMode bool         // Whether to use incognito mode
	browserProxy         string       // The proxy to use
	BrowserSizeH         int          // The horizontal size of the screenshot
	BrowserSizeV         int          // The vertical size of the screenshot
	browserTempDir       string       // The temporary directory to use
	browserUserAgent     string       // The user agent to use. If left empty, a random one will be used
	browserForeground    bool         // Whether to launch the browser in the foreground (headless/not headless)
}

// ParseCliOptions parses the command line options
func ParseCliOptions() (opt *Options, err error) {
	opt = &Options{}
	var targetsRaw string
	var portsRaw string

	flagset := goflags.NewFlagSet()
	flagset.SetDescription("A *blazingly fast*, cross-os cli tool to discover and take automated screenshots of websites using Chromium-based browsers.")
	flagset.CreateGroup("Targets", "Targets",
		flagset.StringVarP(&targetsRaw, "targets", "t", "", "/path/to/urls or a single URL"),
		flagset.StringVarP(&portsRaw, "ports", "p", "default", `ports to scan: "FROM-TO" or "min", "default", "large", "max"`),
		flagset.IntVarP(&opt.PortTimeout, "port-timeout", "pt", 10, "timeout for port checks in seconds"),
		flagset.IntVarP(&opt.Workers, "workers", "w", 10, "amount of concurrect workers"),
	)

	flagset.CreateGroup("Browser", "Browser",
		flagset.StringVarP(&opt.OutputDir, "output-dir", "o", "./screenshots", "output directory for the screenshots"),
		flagset.StringVarP(&opt.browserPath, "chrome", "c", "", "path to the chrome binary"),
		flagset.BoolVarP(&opt.browserIncognitoMode, "incognito", "i", false, "launch chrome in incognito mode"),
		flagset.StringVarP(&opt.browserProxy, "proxy-url", "x", "", "proxy url to use"),
		flagset.IntVarP(&opt.ScreenshotTimeout, "screenshot-timeout", "st", 10, "timeout for the screenshot in seconds"),
		flagset.IntVarP(&opt.BrowserSizeH, "pixel-h", "ph", 1440, "size of the screenshot in pixels (horizontal)"),
		flagset.IntVarP(&opt.BrowserSizeV, "pixel-v", "pv", 800, "size of the screenshot in pixels (vertical)"),
		flagset.StringVarP(&opt.browserTempDir, "temp-dir", "T", "", "directory to store the temporary files"),
		flagset.StringVarP(&opt.browserUserAgent, "user-agent", "ua", "", "override the chrome user agent"),
		flagset.BoolVarP(&opt.browserForeground, "foreground", "fg", false, "launch chrome in foreground"),
	)
	flagset.CreateGroup("Options", "Options",
		flagset.BoolVarP(&logger.DebugOutput, "verbose", "v", false, ""),
	)

	if err = flagset.Parse(); err != nil {
		return nil, err
	}

	opt.Domains, err = parseTargets(targetsRaw,
		func(s string) string { return stripHttp(s) },
		func(s string) bool { return govalidator.IsDNSName(s) },
	)
	if err != nil {
		return opt, err
	}
	logger.Info("Parsed %d target(s)", len(opt.Domains))

	if len(opt.Domains) == 0 {
		return opt, fmt.Errorf("no valid domain(s) parsed")
	}

	opt.Ports, err = parsePorts(portsRaw)
	if err != nil {
		return opt, err
	}
	logger.Info("Parsed %d port(s)", len(opt.Ports))

	_, err = os.Stat(opt.OutputDir)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(opt.OutputDir, 0750)
			if err != nil {
				return nil, fmt.Errorf("creating folder %s failed: %s", opt.OutputDir, err)
			}
		} else {
			return nil, fmt.Errorf("checking folder %s failed: %s", opt.OutputDir, err)
		}
	}

	// Set up the browser
	var l *launcher.Launcher
	if opt.browserPath != "" {
		logger.Info("User wants to use %s", opt.browserPath)
		l = launcher.New().Bin(opt.browserPath)
	} else {
		logger.Info("Checking if Chrome is already present in $PATH")
		if path, isPresent := launcher.LookPath(); isPresent {
			logger.Info("Using %s", path)
			l = launcher.New().Bin(path)
		} else {
			logger.Warning("Chrome not found, using go-rod's Chrome")
			l = launcher.New()
		}
	}

	l = l.
		Headless(!opt.browserForeground).
		UserDataDir(opt.browserTempDir).
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

	if opt.browserIncognitoMode {
		l = l.Set("incognito")
	}
	if opt.browserProxy != "" {
		l = l.Set("proxy-server", opt.browserProxy)
	}
	if opt.browserUserAgent != "" {
		l = l.Set("user-agent", opt.browserUserAgent)
	}
	if os.Geteuid() == 0 {
		l = l.Set("no-sandbox") // Is required under linux & root
	}

	url := l.MustLaunch()
	logger.Info("Connecting to Chrome at %s", url)
	opt.Browser = rod.New().ControlURL(url).MustConnect()

	return opt, nil
}

// SaveScreenshot saves a screenshot to the output directory
func (opt *Options) SaveScreenshot(url string, screenshotBytes []byte) (err error) {
	filePath := generateFilename(opt.OutputDir, url)

	file, err := os.Create(filePath) // #nosec G304 as intended functionality
	if err != nil {
		return fmt.Errorf("creating file %s failed: %s", filePath, err)
	}
	defer file.Close()

	if _, err = file.Write(screenshotBytes); err != nil {
		return fmt.Errorf("writing to file %s failed: %s", filePath, err)
	}

	logger.Info("Saved screenshot of %s to %s", url, filePath)
	return nil
}

// Cleanup closes the Browser and deletes the temp directory
func (opt *Options) Cleanup() (err error) {
	if err = opt.Browser.Close(); err != nil {
		return fmt.Errorf("closing browser failed: %s", err)
	}
	logger.Info("Closed browser")

	if err = os.RemoveAll(opt.browserTempDir); err != nil {
		return fmt.Errorf("removing temp dir %s failed: %s", opt.browserTempDir, err)
	}
	logger.Info("Removed temp dir %s", opt.browserTempDir)

	return nil
}

// parseTargets parses targets from stdin > file via s > string via s
//
// It uses the editTarget function to apply changes to a target. Adjust this to your usecase
//
// It uses the isValidTarget function to check if a target is valiopt.browser Adjust this to your usecase
func parseTargets(s string, editTarget func(string) string, isValidTarget func(string) bool) (targets []string, err error) {
	addTarget := func(t string) {
		t = editTarget(t)
		if isValidTarget(t) && !slices.Contains(targets, t) {
			targets = append(targets, t)
		} else {
			logger.Warning("Skipping %s, not a valid target", t)
		}
	}

	processScanner := func(scanner *bufio.Scanner) error {
		for scanner.Scan() {
			addTarget(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("reading input failed: %w", err)
		}
		return nil
	}

	stdinStat, err := os.Stdin.Stat()
	if err != nil {
		return nil, fmt.Errorf("getting stdin stats failed: %w", err)
	}

	if stdinStat.Mode()&os.ModeNamedPipe != 0 {
		logger.Info("Using stdin, parsing each line as a target")
		if err := processScanner(bufio.NewScanner(os.Stdin)); err != nil {
			return nil, err
		}
	} else {
		if _, err := os.Stat(s); os.IsNotExist(err) {
			logger.Info("File %s not found, treating %s as a single target", s, s)
			addTarget(s)
		} else {
			logger.Info("Treating %s as a path, parsing each line as a target", s)
			file, err := os.Open(s) // #nosec G304 as intended functionality
			if err != nil {
				return nil, fmt.Errorf("opening %s failed: %w", s, err)
			}
			defer file.Close()
			if err := processScanner(bufio.NewScanner(file)); err != nil {
				return nil, err
			}
		}
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("no valid target(s) parsed")
	}

	return targets, nil
}

// stripHttp will remove http(s):// and the trailing slash from a domain.
func stripHttp(domain string) string {
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimSuffix(domain, "/")

	return domain
}

// parsePorts will parse the ports either as an alias or a from-to range
func parsePorts(flag string) (ports []int, err error) {
	switch flag {
	case "min":
		return append(ports, portsMin...), nil
	case "default":
		return append(ports, portsDefault...), nil
	case "large":
		return append(ports, portsLarge...), nil
	case "max":
		return append(ports, portsMax...), nil
	default:
		splitRange := strings.Split(flag, "-")
		if len(splitRange) != 2 {
			return nil, fmt.Errorf("error parsing port range %s", flag)
		}

		fromPort, err := strconv.Atoi(splitRange[0])
		if err != nil {
			return nil, fmt.Errorf("error parsing from port %s: %v", splitRange[0], err)
		}
		toPort, err := strconv.Atoi(splitRange[1])
		if err != nil {
			return nil, fmt.Errorf("error parsing to port %s: %v", splitRange[1], err)
		}

		if fromPort > toPort {
			return nil, fmt.Errorf("from port (%d) cannot be greater than to port (%d)", fromPort, toPort)
		}

		ports = make([]int, 0, toPort-fromPort+1)
		for port := fromPort; port <= toPort; port++ {
			ports = append(ports, port)
		}
		return ports, nil
	}
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

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Filename does not exist
		return filePath
	}

	baseName := safeFilename.String()
	ext := ".png"

	// Iterate until a filename is found that does not exist
	for i := 1; ; i++ {
		newFileName := fmt.Sprintf("%s_%d%s", baseName, i, ext)
		newPath := filepath.Join(dir, newFileName)

		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			// Filename does not exist, return new name
			return newPath
		}
	}
}
