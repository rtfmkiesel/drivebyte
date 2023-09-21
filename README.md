# drivebyte
![GitHub](https://img.shields.io/github/license/rtfmkiesel/drivebyte)

A *blazingly fast*, cross-os cli tool to discover and take automated screenshots of websites using Chromium-based browsers. This was done with automated content discovery & bug bounty in mind. This was heavily inspired and is basically a minified, zero-dependency version of [michenriksen/aquatone](https://github.com/michenriksen/aquatone). 

![](demo.gif)

For chaining, `drivebyte` will output successfully screenshotted URLs to `stdout` and verbose as well as error messages to `stderr` which makes the output easily parsable. 

## Usage
```
drivebyte [OPTIONS]

Examples:
    drivebyte -u "domain.tld"
    drivebyte -f domains.txt
    echo "domain.tld" | drivebyte
    cat domains.txt | drivebyte

Options:
    -u,  --url               <string>    A single domain/URL to enumerate and screenshot
    -f,  --file              <string>    Path to a file containing one URL to screenshot per line
    -o,  --output-dir        <string>    Path to the output folder for screenshots (default: ./screenshots)
    
    -p,  --ports             <string>    Ports to scan: "from-to" or "min", "default", "large", "max"
    -pt, --port-timeout      <int>       Timeout for port checks in seconds (default: 3)

    -c,  --chrome            <string>    Path to the chrome binary (default: autodetect)
    -i,  --incognito                     Use incognito mode (default: false)
    -x,  --proxy-url         <string>    Proxy URL for the webbrowser "http://user:passwd@host:port"
    -st, --screen-timeout    <int>       Timeout for taking screenshots in seconds (default: 10)
    -ph, --pixel-h           <int>       Amount of horizontal pixels for the browser window (default: 1440)
    -pv, --pixel-v           <int>       Amount of vertical pixels for the browser window (default: 800)
    -t,  --temp-dir          <string>    Parent directory for browser cache (default: OSTEMP)
    -ua, --user-agent        <string>    User-Agent for the webbrowser (default: random)

    -w,  --workers           <int>       Amount of "threads" aka. browsers open at the same time (default: 3)
    -v,  --verbose                       Enable verbose output (default: false)
    -h,  --help                          Prints this text
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
go mod tidy
go build -ldflags="-s -w" .
```
