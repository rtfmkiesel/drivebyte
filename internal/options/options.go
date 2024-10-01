package options

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/projectdiscovery/goflags"

	"github.com/rtfmkiesel/drivebyte/internal/logger"
	"github.com/rtfmkiesel/drivebyte/internal/screenshoter"
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
	Domains     []string // The domains to test for screenshots
	Ports       []int    // The ports to test for potential web servers
	PortTimeout int      // The timeout for port checks
	Workers     int      // The number of concurrent workers
}

// Parse parses the command line options
func Parse() (opt *Options, s *screenshoter.Screenshoter, err error) {
	opt = &Options{}
	screenshotOpt := &screenshoter.Options{}

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
		flagset.StringVarP(&screenshotOpt.OutputDir, "output-dir", "o", "./screenshots", "output directory for the screenshots"),
		flagset.StringVarP(&screenshotOpt.ChromePath, "chrome", "c", "", "path to the chrome binary"),
		flagset.BoolVarP(&screenshotOpt.Incognito, "incognito", "i", false, "launch chrome in incognito mode"),
		flagset.StringVarP(&screenshotOpt.Proxy, "proxy-url", "x", "", "proxy url to use"),
		flagset.IntVarP(&screenshotOpt.Timeout, "screenshot-timeout", "st", 10, "timeout for the screenshot in seconds"),
		flagset.IntVarP(&screenshotOpt.BrowserSizeH, "pixel-h", "ph", 1440, "size of the screenshot in pixels (horizontal)"),
		flagset.IntVarP(&screenshotOpt.BrowserSizeV, "pixel-v", "pv", 800, "size of the screenshot in pixels (vertical)"),
		flagset.StringVarP(&screenshotOpt.TempPath, "temp-dir", "T", "", "directory to store the temporary files (default mktemp)"),
		flagset.StringVarP(&screenshotOpt.UserAgent, "user-agent", "ua", "", "override the chrome user agent"),
		flagset.BoolVarP(&screenshotOpt.Foreground, "foreground", "fg", false, "launch chrome in foreground"),
	)
	flagset.CreateGroup("Options", "Options",
		flagset.BoolVarP(&logger.ShowDebugOutput, "verbose", "v", false, ""),
	)

	if err := flagset.Parse(); err != nil {
		return nil, nil, err
	}

	opt.Domains, err = parseTargets(targetsRaw)
	if err != nil {
		return nil, nil, err
	}
	logger.Debug("Parsed %d target(s)", len(opt.Domains))

	if len(opt.Domains) == 0 {
		return nil, nil, fmt.Errorf("no valid domain(s) parsed")
	}

	opt.Ports, err = parsePorts(portsRaw)
	if err != nil {
		return nil, nil, err
	}
	logger.Debug("Parsed %d port(s)", len(opt.Ports))

	s, err = screenshoter.NewScreenshoter(screenshotOpt)
	if err != nil {
		return nil, nil, err
	}

	return opt, s, nil
}

// parseTargets parses targets from stdin > file via s > string via s
func parseTargets(s string) (targets []string, err error) {
	addTarget := func(t string) {
		t = strings.TrimPrefix(t, "http://")
		t = strings.TrimPrefix(t, "https://")
		t = strings.TrimSuffix(t, "/")

		if govalidator.IsURL(t) && !slices.Contains(targets, t) {
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
		logger.Debug("Using stdin, parsing each line as a target")
		if err := processScanner(bufio.NewScanner(os.Stdin)); err != nil {
			return nil, err
		}
	} else {
		if _, err := os.Stat(s); os.IsNotExist(err) {
			logger.Debug("File %s not found, treating %s as a single target", s, s)
			addTarget(s)

		} else {
			logger.Debug("Treating %s as a path, parsing each line as a target", s)
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
