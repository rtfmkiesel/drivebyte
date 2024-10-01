package webdiscovery

import (
	"fmt"
	"net/http"
	"time"

	"github.com/rtfmkiesel/drivebyte/internal/logger"
)

var (
	securePorts = map[int]bool{
		443:   true,
		832:   true,
		981:   true,
		1010:  true,
		1311:  true,
		2083:  true,
		2087:  true,
		2095:  true,
		2096:  true,
		4712:  true,
		7000:  true,
		8172:  true,
		8243:  true,
		8333:  true,
		8443:  true,
		8834:  true,
		9443:  true,
		12443: true,
		18091: true,
		18092: true,
	}
)

// GenerateUrls returns []string containing all possible URLs for a given []string and []int for hte ports
func GenerateUrls(domains []string, ports []int) (urls []string) {
	for _, domain := range domains {
		for _, port := range ports {
			protocol := "http"
			if securePorts[port] {
				protocol = "https"
			}

			if port == 80 || port == 443 { // Do not add :80 and :443 since those are the default ports
				urls = append(urls, fmt.Sprintf("%s://%s", protocol, domain))
			} else {
				urls = append(urls, fmt.Sprintf("%s://%s:%d", protocol, domain, port))
			}
		}
	}

	logger.Debug("Generated a total of %d possible URLs for %d domain(s)", len(urls), len(domains))
	return urls
}

// IsReachable will test connectivity to a web server via a GET request.
func IsReachable(url string, timeout int) bool {
	logger.Debug("Checking for website at %s", url)

	httpClient := http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	response, err := httpClient.Get(url)
	if err != nil || len(response.Header) == 0 {
		return false
	}
	defer response.Body.Close()

	logger.Debug("Website found at %s", url)
	return true
}
