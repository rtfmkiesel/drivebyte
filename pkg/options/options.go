// options handels the command line options
package options

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"drivebyte/pkg/logger"
)

var (
	// https://github.com/michenriksen/aquatone/blob/master/core/ports.go
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

type Options struct {
	Domains           []string
	OutputDir         string
	Ports             []int
	PortTimeout       int
	ChromePath        string
	IncognitoMode     bool
	Proxy             string
	ScreenshotTimeout int
	SizeH             int
	SizeV             int
	TempDir           string
	UserAgent         string
	Workers           int
}

// Parse() will parse all the command line options and get the Urls from either stdin or the inputfile
func Parse() (opt Options, err error) {
	var flagInputfile string // Seperate since this flag is only needed in Parse()
	var flagPorts string     // Seperate since this flag is only needed in Parse()

	// Fun with flags
	flag.StringVar(&flagInputfile, "f", "", "")
	flag.StringVar(&flagInputfile, "file", "", "")
	flag.StringVar(&opt.OutputDir, "o", "./screenshots", "")
	flag.StringVar(&opt.OutputDir, "output-dir", "./screenshots", "")
	flag.StringVar(&flagPorts, "p", "default", "")
	flag.StringVar(&flagPorts, "ports", "default", "")
	flag.IntVar(&opt.PortTimeout, "pt", 3, "")
	flag.IntVar(&opt.PortTimeout, "port-timeout", 3, "")
	flag.StringVar(&opt.ChromePath, "c", "", "")
	flag.StringVar(&opt.ChromePath, "chrome", "", "")
	flag.BoolVar(&opt.IncognitoMode, "i", false, "")
	flag.BoolVar(&opt.IncognitoMode, "incognito", false, "")
	flag.StringVar(&opt.Proxy, "x", "", "")
	flag.StringVar(&opt.Proxy, "proxy-url", "", "")
	flag.IntVar(&opt.ScreenshotTimeout, "st", 10, "")
	flag.IntVar(&opt.ScreenshotTimeout, "screenshot-timeout", 10, "")
	flag.IntVar(&opt.SizeH, "ph", 1440, "")
	flag.IntVar(&opt.SizeH, "pixel-h", 1440, "")
	flag.IntVar(&opt.SizeV, "pv", 800, "")
	flag.IntVar(&opt.SizeV, "pixel-v", 800, "")
	flag.StringVar(&opt.TempDir, "t", "", "")
	flag.StringVar(&opt.TempDir, "temp-dir", "", "")
	flag.StringVar(&opt.UserAgent, "ua", "", "")
	flag.StringVar(&opt.UserAgent, "user-agent", "", "")
	flag.IntVar(&opt.Workers, "w", 3, "")
	flag.IntVar(&opt.Workers, "workers", 3, "")
	flag.BoolVar(&logger.Verbose, "v", false, "")
	flag.BoolVar(&logger.Verbose, "verbose", false, "")
	flag.Usage = func() { usage() }
	flag.Parse()

	// Parse ports
	opt.Ports, err = parsePorts(flagPorts)
	if err != nil {
		return opt, err
	}

	// Parse domains via file or stdin
	opt.Domains, err = parseDomains(flagInputfile)
	if err != nil {
		return opt, err
	}

	// Check dirs and set up missing ones
	err = checkDirs(&opt)
	if err != nil {
		return opt, err
	}

	return opt, nil
}

// CreateDirIfNotExist() will check if a directory exists and create it if not.
// Additionally it will make sure that the newly created directory or the already existing one is writeable.
func CreateDirIfNotExist(dir string) (err error) {
	_, err = os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			// Folder does not exist, create
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				return fmt.Errorf("creating folder '%s' failed: '%s'", dir, err)
			}
		} else {
			return fmt.Errorf("checking folder '%s' failed: '%s'", dir, err)
		}
	}

	// Folder exists, check if it's writable
	filePath := filepath.Join(dir, "fsociety00.dat")
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("writing to folder '%s' failed: '%s'", dir, err)
	}
	err = file.Close()
	if err != nil {
		return fmt.Errorf("closing temp file failed: '%s'", err)
	}

	err = os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("deleting temp file failed: '%s'", err)
	}

	logger.Info("Created directory '%s'", dir)
	return nil
}

// parsePorts() will parse the ports. Either an alias is supported or a from-to range
func parsePorts(flag string) (ports []int, err error) {
	switch flag {
	case "min":
		ports = append(ports, portsMin...)
	case "default":
		ports = append(ports, portsDefault...)
	case "large":
		ports = append(ports, portsLarge...)
	case "max":
		ports = append(ports, portsMax...)
	default:
		// Port range
		splitRange := strings.Split(flag, "-")
		if len(splitRange) != 2 {
			return nil, fmt.Errorf("error parsing port range '%s'", flag)
		}

		// From string to int
		fromPort, err := strconv.Atoi(splitRange[0])
		if err != nil {
			return nil, fmt.Errorf("error parsing port range '%s'", flag)
		}
		toPort, err := strconv.Atoi(splitRange[1])
		if err != nil {
			return nil, fmt.Errorf("error parsing port range '%s'", flag)
		}

		for port := fromPort; port <= toPort; port++ {
			ports = append(ports, port)
		}
	}

	return ports, nil
}

// parseDomains() will create a list of domains based on either stdin or the provided flag.
// If the flag for the input file is "", stdin will be use
func parseDomains(flag string) (domains []string, err error) {
	if flag != "" {
		// Get domain via file

		// Check if the file exists.
		_, err := os.Stat(flag)
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("inputfile '%s' does not exist", flag)
		}

		// Open file
		file, err := os.Open(flag)
		if err != nil {
			return nil, fmt.Errorf("opening inputfile '%s' failed: '%s'", flag, err)
		}
		defer file.Close()

		// Read line by line
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			domain := scanner.Text()
			// Check if supplied domain is a valid URL to avoid later errors
			if !isDomain(domain) {
				logger.Error("ignoring '%s', not a valid URL", domain)
				continue
			} else {
				domains = append(domains, domain)
			}
		}

		// Check for scanner errors
		err = scanner.Err()
		if err != nil {
			return nil, fmt.Errorf("reading inputfile '%s' failed: '%s'", flag, err)
		}

	} else {
		// Get urls via stdin

		// Check that stdin != empty
		stdinStat, err := os.Stdin.Stat()
		if err != nil {
			return nil, fmt.Errorf("reading of stdin failed: '%s'", err)
		}
		if stdinStat.Mode()&os.ModeNamedPipe == 0 {
			return nil, fmt.Errorf("stdin was empty and no inputfile has been specified")
		}

		// Read from stdin
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			domain := scanner.Text()
			// Check if supplied domain is a valid URL to avoid later errors
			if !isDomain(domain) {
				continue
			} else {
				domains = append(domains, domain)
			}
		}

		// Check for scanner errors
		err = scanner.Err()
		if err != nil {
			return nil, fmt.Errorf("reading inputfile '%s' failed: '%s'", flag, err)
		}
	}

	return domains, nil
}

// isDomain() returns true if a domain is a valid domain name
func isDomain(domain string) bool {
	// https://github.com/asaskevich/govalidator/blob/master/patterns.go
	pattern := `^([a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62}){1}(\.[a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62})*[\._]?$`
	regex := regexp.MustCompile(pattern)
	return regex.MatchString(domain)
}

// checkDirs() will check the paths which the user supplied (browser, output)
func checkDirs(opt *Options) (err error) {
	// Check browserPath
	if opt.ChromePath != "" {
		if _, err := os.Stat(opt.ChromePath); os.IsNotExist(err) {
			// User supplied path not accessible
			return fmt.Errorf("accessing '%s' failed: '%s'", opt.ChromePath, err)
		}
	} else {
		// Check the default paths
		opt.ChromePath, err = locateBrowser()
		if err != nil {
			// No Installation found
			return err
		}
	}
	logger.Info("Using browser '%s'", opt.ChromePath)

	// Check imgOutputDir
	if opt.OutputDir == "" {
		opt.OutputDir, err = filepath.Abs("./screenshots")
		if err != nil {
			return fmt.Errorf("getting abosulte path of './screenshots' failed: '%s'", err)
		}
	} else {
		opt.OutputDir, err = filepath.Abs(opt.OutputDir)
		if err != nil {
			return fmt.Errorf("getting abosulte path of '%s' failed: '%s'", opt.OutputDir, err)
		}
	}
	err = CreateDirIfNotExist(opt.OutputDir)
	if err != nil {
		return err
	}
	logger.Info("Saving screenshot(s) to '%s'", opt.OutputDir)

	return nil
}

// locateBrowser() looks in default paths for the browser.
func locateBrowser() (browserPath string, err error) {
	browserPaths := []string{}
	logger.Info("Locating browser installation")

	// https://github.com/go-rod/rod/blob/main/lib/launcher/browser.go
	switch runtime.GOOS {
	case "windows":
		browserPaths = []string{
			"C:/Program Files/Google/Chrome/Application/chrome.exe",
			"C:/Program Files (x86)/Google/Chrome/Application/chrome.exe",
			"C:/Program Files/Chromium/Application/chrome.exe",
			"C:/Program Files (x86)/Chromium/Application/chrome.exe",
			"C:/Program Files/Microsoft/Edge/Application/msedge.exe",
			"C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe",
		}
	case "linux", "openbsd":
		browserPaths = []string{
			"chrome",
			"google-chrome",
			"/usr/bin/google-chrome",
			"microsoft-edge",
			"/usr/bin/microsoft-edge",
			"chromium",
			"chromium-browser",
			"/usr/bin/google-chrome-stable",
			"/usr/bin/chromium",
			"/usr/bin/chromium-browser",
			"/snap/bin/chromium",
			"/data/data/com.termux/files/usr/bin/chromium-browser",
		}
	case "darwin":
		browserPaths = []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
			"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
			"/Applications/Google Chrome Canary.app/Contents/MacOS/Google Chrome Canary",
			"/usr/bin/google-chrome-stable",
			"/usr/bin/google-chrome",
			"/usr/bin/chromium",
			"/usr/bin/chromium-browser",
		}
	}

	for _, path := range browserPaths {
		_, err := os.Stat(path)
		if err == nil {
			// Browser found
			return path, nil
		}
	}

	// No path matched
	return "", fmt.Errorf("no browser found, please specify path")
}

// usage() will print the help text
func usage() {
	fmt.Printf(`drivebyte [OPTIONS]

Examples:
    echo "domain.tld" | drivebyte
    cat domains.txt | drivebyte
    drivebyte -f domains.txt

Options:
    -f, --file              <string>    Path to a file containing one URL to screenshot per line
    -o, --output-dir        <string>    Path to the output folder for screenshots (default: ./screenshots)
    
    -p, --ports             <string>    Ports to scan: "from-to" or "min", "default", "large", "max"
    -pt, --port-timeout     <int>       Timeout for port checks in seconds (default: 3)

    -c, --chrome            <string>    Path to the chrome binary. If left empty, autodetect is used
    -i, --incognito                     Use incognito mode. (default: false)
    -x, --proxy-url         <string>    Proxy URL for the webbrowser. "http://user:passwd@host:port"
    -st, --screen-timeout   <int>       Timeout for taking screenshots in seconds. (default: 10)
    -ph, --pixel-h          <int>       Amount of horizontal pixels for the browser window. (default: 1440)
    -pv, --pixel-v          <int>       Amount of vertical pixels for the browser window. (default: 800)
    -t, --temp-dir          <string>    Parent directory for browser cache. (default: .)
    -ua, --user-agent       <string>    User-Agent for the webbrowser. (default: random)

    -w, --workers           <int>       Amount of "threads" aka. browsers open at the same time. (default: 3)
    -v, --verbose                       Enable verbose output. (default: false)
    -h, --help                          Prints this text

`)
}
