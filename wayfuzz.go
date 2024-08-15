package main

import (
        "bufio"
        "fmt"
        "net/http"
        "os"
        "regexp"
        "strings"
)
func main() {
        scanner := bufio.NewScanner(os.Stdin)
        for scanner.Scan() {
                domain := scanner.Text()
                urls, err := fetchWaybackURLs(domain)
                if err != nil {
                        fmt.Fprintf(os.Stderr, "Error fetching URLs for domain %s: %v\n", domain, err)
                        continue
                }
                for _, url := range urls {
                        trimmedURL := trimURL(url, domain)
                        if trimmedURL != "" {
                                fmt.Println(trimmedURL)
                        }
                }
        }

        if err := scanner.Err(); err != nil {
                fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
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
        re := regexp.MustCompile(`^https?://([a-zA-Z0-9_-]+\.)*` + regexp.QuoteMeta(domain))
        return re.ReplaceAllString(url, "")
}
