# wayfuzz

is a fast and efficient tool for creating wordlists from historical URLs fetched via the Wayback Machine. You can use it with tools like ffuf for web fuzzing and other security testing tasks.

## Features

    Concurrency: Make multiple requests simultaneously for faster processing.
    URL Filtering: Exclude specific URL patterns using regex.
    Path Separation: Optionally split URL paths into distinct components.
    Status Code: specify a comma-separated list of status codes (e.g., -mc 200,403).


## Installation

`go install -v github.com/Vulnpire/wayfuzz@latest`

## Or build from the source

Clone the repo

`git clone https://github.com/Vulnpire/wayfuzz`

And build

`go build -o wayfuzz wayfuzz.go`

This will create an executable named `wayfuzz`.

## Usage

You can use wayfuzz by piping in a list of domains via `stdin`:

`cat domains.txt | wayfuzz [options]`

## Options

    -c <int>: Set the number of concurrent requests (default: 10).
    -x <regex>: Exclude URLs matching the regex pattern (e.g., .jpg|.png).
    -sed: Split the URL paths by / and output each component separately.
    -mc <codes>: Filter URLs by status codes (comma-separated list, e.g., 200,403).

## Example Commands
### Basic Usage

Exclude URLs that end in .jpg or .png:

`cat domains.txt | wayfuzz -c 50`

Exclude Specific URL Patterns:

`cat domains.txt | wayfuzz -c 50 -x ".jpg|.png"`

Separate URL Paths by `/`

`cat domains.txt | wayfuzz -c 50 -sed`

Filter by Status Codes

`cat domains.txt | wayfuzz -c 50 -mc 200,403`

## Using with `ffuf`

`ffuf` is a web fuzzing tool that can be combined with `wayfuzz` for discovering hidden files, directories, and parameters on a web server.

`cat domains.txt | wayfuzz -c 50 | ffuf -u https://target.com/FUZZ -w -`

If you want to fuzz URL parameters, you can generate a wordlist of all unique URL components:

`cat domains.txt | wayfuzz -c 50 -sed | ffuf -u https://target.com/path?FUZZ=value -w -`

## IP fuzzing

Creating the wordlist:

`echo "hackerone.com" | wayfuzz -c 300 -mc 200 -sed -x ".jpg|.png|.jpeg|..." | anew wordlist.txt`

Getting the IP addresses from Shodan:

`echo "hackerone.com" | `[sXtract](http://github.com:443/Vulnpire/sXtract)` | anew ips.txt`

Fuzzing the IPs:

`cat ips.txt | xargs -I@ sh -c 'ffuf -w ./wordlist.txt -u @/FUZZ -mc 200 -c -recursion -recursion-depth 5 -ac -t 300'` Or just use Axiom to fuzz quickly.
