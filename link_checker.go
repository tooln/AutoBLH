package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// List of extensions to exclude
var extensionsToExclude = []string{".png", ".gif", ".webp", ".bmp", ".pdf", ".psd", ".jpg", ".jpeg", ".tiff", ".eps", ".ai", ".raw", ".indd", ".css", ".js"}

// Check if a link should be excluded based on its extension
func shouldExclude(link string) bool {
	u, err := url.Parse(link)
	if err != nil {
		return true
	}
	path := strings.ToLower(u.Path)
	for _, ext := range extensionsToExclude {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}

// Check the status of a link
func checkLink(link string, wg *sync.WaitGroup, ch chan string) {
	defer wg.Done()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Head(link)
	if err == nil && resp.StatusCode == http.StatusOK {
		ch <- link
	} else {
		ch <- ""
	}
}

func main() {
	// Parse command-line arguments
	linksFile := flag.String("links", "", "Path to the file containing the list of links.")
	outputFile := flag.String("output", "", "Path to the output file where the results will be saved.")
	flag.Parse()

	if *linksFile == "" || *outputFile == "" {
		fmt.Println("Please provide both the input and output file paths.")
		flag.Usage()
		return
	}

	// Read the links from the file
	file, err := os.Open(*linksFile)
	if err != nil {
		fmt.Printf("Error opening links file: %v\n", err)
		return
	}
	defer file.Close()

	var links []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		link := strings.TrimSpace(scanner.Text())
		if link != "" && !shouldExclude(link) {
			links = append(links, link)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading links file: %v\n", err)
		return
	}

	// Check the links concurrently
	var wg sync.WaitGroup
	ch := make(chan string, len(links))
	for _, link := range links {
		wg.Add(1)
		go checkLink(link, &wg, ch)
	}

	// Wait for all goroutines to finish
	go func() {
		wg.Wait()
		close(ch)
	}()

	// Collect the alive links
	var aliveLinks []string
	for link := range ch {
		if link != "" {
			aliveLinks = append(aliveLinks, link)
		}
	}

	// Write the alive links to the output file
	output, err := os.Create(*outputFile)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		return
	}
	defer output.Close()

	writer := bufio.NewWriter(output)
	for _, link := range aliveLinks {
		writer.WriteString(link + "\n")
	}
	writer.Flush()

	fmt.Printf("Alive links saved to %s\n", *outputFile)
}
  
