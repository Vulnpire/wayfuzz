package main

import (
        "bufio"
        "flag"
        "fmt"
        "net/http"
        "os"
        "regexp"
        "sort"
        "strconv"
        "strings"
        "sync"
)

func main() {
        // Command-line flags
        concurrency := flag.Int("c", 10, "Number of concurrent requests")
        excludePattern := flag.String("x", "", "Regex pattern to exclude certain URLs (e.g., .jpg|.png)")
        separateSlash := flag.Bool("sed", false, "Separate URL paths by '/'")
        mc := flag.String("mc", "", "Comma-separated list of status codes to include (e.g., 200,403)")
        flag.Parse()

        // Regex for exclusion
        var excludeRegex *regexp.Regexp
        if *excludePattern != "" {
                excludeRegex = regexp.MustCompile(*excludePattern)
        }

        // Parse the list of status codes
        var statusCodes map[int]struct{}
        if *mc != "" {
                statusCodes = parseStatusCodes(*mc)
        }

        // Create a channel to handle concurrency
        jobs := make(chan string)
        results := make(chan []string)

        // Create a WaitGroup to wait for all goroutines to finish
        var wg sync.WaitGroup

        // Start workers
        for i := 0; i < *concurrency; i++ {
                wg.Add(1)
                go worker(jobs, results, &wg, excludeRegex, statusCodes, *separateSlash)
        }

        // Start a goroutine to close the results channel when done
        go func() {
                wg.Wait()
                close(results)
        }()

        // Scan input domains and send them to jobs channel
        go func() {
                scanner := bufio.NewScanner(os.Stdin)
                for scanner.Scan() {
                        domain := scanner.Text()
                        jobs <- domain
                }
                close(jobs)
        }()

        // Collect and deduplicate results
        urlSet := make(map[string]struct{})
        for urls := range results {
                for _, url := range urls {
                        cleanedURL := strings.TrimSpace(url)
                        if cleanedURL != "" {
                                urlSet[cleanedURL] = struct{}{}
                        }
                }
        }

        // Sort and print unique results
        uniqueUrls := make([]string, 0, len(urlSet))
        for url := range urlSet {
                uniqueUrls = append(uniqueUrls, url)
        }
        sort.Strings(uniqueUrls)
        for _, url := range uniqueUrls {
                fmt.Println(url)
        }
}

func worker(jobs <-chan string, results chan<- []string, wg *sync.WaitGroup, excludeRegex *regexp.Regexp, statusCodes map[int]struct{}, separateSlash bool) {
        defer wg.Done()
        for domain := range jobs {
                urls, err := fetchWaybackURLs(domain, statusCodes)
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

func fetchWaybackURLs(domain string, statusCodes map[int]struct{}) ([]string, error) {
        waybackURL := fmt.Sprintf("http://web.archive.org/cdx/search/cdx?collapse=urlkey&url=*.%s", domain)
        if statusCodes != nil {
                waybackURL += "&filter=statuscode:" + buildStatusCodeFilter(statusCodes)
        }
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
                        statusCode, _ := strconv.Atoi(fields[4])
                        if _, ok := statusCodes[statusCode]; ok || statusCodes == nil {
                                urls = append(urls, fields[2])
                        }
                }
        }

        if err := scanner.Err(); err != nil {
                return nil, err
        }

        return urls, nil
}

func trimURL(url, domain string) string {
        // Regex to remove the scheme (http, https), any ports, and domain (case insensitive)
        re := regexp.MustCompile(`(?i)^https?://([a-zA-Z0-9_-]+\.)*` + regexp.QuoteMeta(strings.ToLower(domain)) + `(:\d+)?`)
        return re.ReplaceAllString(url, "")
}

func parseStatusCodes(mc string) map[int]struct{} {
        codes := strings.Split(mc, ",")
        statusCodes := make(map[int]struct{}, len(codes))
        for _, code := range codes {
                if c, err := strconv.Atoi(code); err == nil {
                        statusCodes[c] = struct{}{}
                }
        }
        return statusCodes
}

func buildStatusCodeFilter(statusCodes map[int]struct{}) string {
        var codes []string
        for code := range statusCodes {
                codes = append(codes, strconv.Itoa(code))
        }
        return strings.Join(codes, ",")
}
