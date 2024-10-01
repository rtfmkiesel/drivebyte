# drivebyte
![GitHub Repo stars](https://img.shields.io/github/stars/rtfmkiesel/drivebyte) ![GitHub](https://img.shields.io/github/license/rtfmkiesel/drivebyte)

A *blazingly fast*, cross-os cli tool to discover and take automated screenshots of websites using Chromium-based browsers. This was done with automated content discovery & bug bounty in mind. This was heavily inspired and is basically a minified version of [michenriksen/aquatone](https://github.com/michenriksen/aquatone). 

![](demo.gif)

For chaining, `drivebyte` will output successfully screenshotted URLs to `stdout` and verbose as well as error messages to `stderr` which makes the output easily parsable. 

## Usage
```
Usage:
  drivebyte [flags]

Flags:
TARGETS:
   -t, -targets string     /path/to/urls or a single URL
   -p, -ports string       ports to scan: "FROM-TO" or "min", "default", "large", "max" (default "default")
   -pt, -port-timeout int  timeout for port checks in seconds (default 10)
   -w, -workers int        amount of concurrect workers (default 10)

BROWSER:
   -o, -output-dir string        output directory for the screenshots (default "./screenshots")
   -c, -chrome string            path to the chrome binary
   -i, -incognito                launch chrome in incognito mode
   -x, -proxy-url string         proxy url to use
   -st, -screenshot-timeout int  timeout for the screenshot in seconds (default 10)
   -ph, -pixel-h int             size of the screenshot in pixels (horizontal) (default 1440)
   -pv, -pixel-v int             size of the screenshot in pixels (vertical) (default 800)
   -T, -temp-dir string          directory to store the temporary files
   -ua, -user-agent string       override the chrome user agent
   -fg, -foreground              launch chrome in foreground

OPTIONS:
   -v, -verbose
```

## Installation
### Binaries
Download the prebuilt binaries [here](https://github.com/rtfmkiesel/drivebyte/releases).

### With Go
```bash
go install github.com/rtfmkiesel/drivebyte@latest
```

### Build from source
```bash
git clone https://github.com/rtfmkiesel/drivebyte
cd drivebyte

make
# or
go mod tidy
go build -ldflags="-s -w" .
```
