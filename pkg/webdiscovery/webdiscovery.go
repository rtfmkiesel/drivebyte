// webdiscovery handles to discovery of full web URLs from provided domains
package webdiscovery

import (
	"fmt"
	"net/http"
	"time"

	"github.com/rtfmkiesel/drivebyte/pkg/logger"
	"github.com/rtfmkiesel/drivebyte/pkg/options"
)

var (
	securePorts = []int{443, 832, 981, 1010, 1311, 2083, 2087, 2095, 2096, 4712,
		7000, 8172, 8243, 8333, 8443, 8834, 9443, 12443, 18091, 18092}
)

// Generate() returns []string containing all possible URLs for a given []string of domains within the options.Options struct.
func Generate(opt options.Options) (urls []string) {
	for _, domain := range opt.Domains {
		for _, port := range opt.Ports {
			protocol := "http"
			if isSecurePort(port) {
				protocol = "https"
			}

			if port == 80 || port == 443 {
				// Do not add :80 and :443 since those are the default ports
				urls = append(urls, fmt.Sprintf("%s://%s", protocol, domain))
			} else {
				urls = append(urls, fmt.Sprintf("%s://%s:%d", protocol, domain, port))
			}
		}
	}

	logger.Info("Generated a total of %d possible web servers for %d domain(s)", len(urls), len(opt.Domains))
	return urls
}

// IsOpen() will test connectivity to a web server via a GET request.
func IsOpen(url string, timeout int) bool {
	logger.Info("Testing '%s'", url)

	httpClient := http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	response, err := httpClient.Get(url)
	if err != nil || len(response.Header) == 0 {
		return false
	}
	defer response.Body.Close()

	logger.Success("Webserver found at '%s'", url)
	return true
}

// isSecurePort will return true if a given port is in the list of secure ports. This is required to set the protocol inside the URL correctly.
func isSecurePort(port int) bool {
	for _, p := range securePorts {
		if p == port {
			return true
		}
	}

	return false
}
