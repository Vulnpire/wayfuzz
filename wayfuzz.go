package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
)

func main() {
	concurrency := flag.Int("c", 10, "Number of concurrent requests")
	excludePattern := flag.String("x", "", "Regex pattern to exclude certain URLs (e.g., .jpg|.png)")
	separateSlash := flag.Bool("sed", false, "Separate URL paths by '/'")
	flag.Parse()

	var excludeRegex *regexp.Regexp
	if *excludePattern != "" {
		excludeRegex = regexp.MustCompile(*excludePattern)
	}

	jobs := make(chan string)
	results := make(chan []string)

	var wg sync.WaitGroup

	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go worker(jobs, results, &wg, excludeRegex, *separateSlash)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			domain := scanner.Text()
			jobs <- domain
		}
		close(jobs)
	}()
	urlSet := make(map[string]struct{})
	for urls := range results {
		for _, url := range urls {
			cleanedURL := strings.TrimSpace(url)
			if cleanedURL != "" {
				urlSet[cleanedURL] = struct{}{}
			}
		}
	}

	uniqueUrls := make([]string, 0, len(urlSet))
	for url := range urlSet {
		uniqueUrls = append(uniqueUrls, url)
	}
	sort.Strings(uniqueUrls)
	for _, url := range uniqueUrls {
		fmt.Println(url)
	}
}

func worker(jobs <-chan string, results chan<- []string, wg *sync.WaitGroup, excludeRegex *regexp.Regexp, separateSlash bool) {
	defer wg.Done()
	for domain := range jobs {
		urls, err := fetchWaybackURLs(domain)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching URLs for domain %s: %v\n", domain, err)
			continue
		}
		var filteredUrls []string
		for _, url := range urls {
			trimmedURL := trimURL(url, domain)
			if trimmedURL != "" && (excludeRegex == nil || !excludeRegex.MatchString(trimmedURL)) {
				if separateSlash {
					parts := strings.Split(trimmedURL, "/")
					for _, part := range parts {
						cleanedPart := strings.TrimSpace(part)
						if cleanedPart != "" {
							filteredUrls = append(filteredUrls, cleanedPart)
						}
					}
				} else {
					filteredUrls = append(filteredUrls, strings.TrimSpace(trimmedURL))
				}
			}
		}
		results <- filteredUrls
	}
}

func fetchWaybackURLs(domain string) ([]string, error) {
	waybackURL := fmt.Sprintf("http://web.archive.org/cdx/search/cdx?collapse=urlkey&url=*.%s", domain)
	resp, err := http.Get(waybackURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var urls []string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) > 2 {
			urls = append(urls, fields[2])
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}

func trimURL(url, domain string) string {
	re := regexp.MustCompile(`(?i)^https?://([a-zA-Z0-9_-]+\.)*` + regexp.QuoteMeta(strings.ToLower(domain)) + `(:\d+)?`)
	return re.ReplaceAllString(url, "")
}
