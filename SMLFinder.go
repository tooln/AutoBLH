package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

var filterStrings = []string{
	"facebook.com", "twitter.com", "instagram.com", "linkedin.com", "youtube.com",
	"snapchat.com", "pinterest.com", "reddit.com", "tiktok.com", "tumblr.com",
	"whatsapp.com", "wechat.com", "telegram.org", "vimeo.com", "medium.com",
	"periscope.tv", "twitch.tv", "discord.com", "mastodon.social", "bandcamp.com",
	"vk.com", "github.com", "livejournal.com", "xing.com", "t.me",
	"linkin.bio", "threads.com", "bit.ly", "tinyurl.com", "github.io",
	"linktr.ee", "onelink.bio",
}

// Download the source code of the webpage
func downloadSourceCode(url string, wg *sync.WaitGroup, ch chan<- string) {
	defer wg.Done()

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error downloading %s: %v\n", url, err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body for %s: %v\n", url, err)
		return
	}

	filename := strings.ReplaceAll(url, "/", "_") + ".html"
	err = ioutil.WriteFile(filename, body, 0644)
	if err != nil {
		fmt.Printf("Error writing to file %s: %v\n", filename, err)
		return
	}

	fmt.Printf("Source code downloaded successfully as %s\n", filename)
	ch <- filename
}

// Search and highlight links in the downloaded source code
func searchAndHighlightLinks(filename string, filterStrings []string, url string, outputChan, saveChan chan<- string) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", filename, err)
		return
	}

	found := false
	var filteredLinks []string
	for _, filter := range filterStrings {
		re := regexp.MustCompile(`https?://[^\s]*` + regexp.QuoteMeta(filter) + `[^\s]*`)
		links := re.FindAllString(string(data), -1)
		if len(links) > 0 {
			found = true
			for _, link := range links {
				fmt.Println(link)
				filteredLinks = append(filteredLinks, link)
				outputChan <- link
			}
		}
	}

	if found && saveChan != nil {
		saveChan <- fmt.Sprintf("Downloaded URL: %s\nFiltered links: %s\n", url, strings.Join(filteredLinks, ", "))
	} else {
		fmt.Printf("No filtered links found in %s\n", filename)
	}
}

// Delete all downloaded source code files
func deleteDownloadedFiles() {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		fmt.Printf("Error reading directory: %v\n", err)
		return
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".html") {
			err = os.Remove(file.Name())
			if err != nil {
				fmt.Printf("Error deleting file %s: %v\n", file.Name(), err)
			}
		}
	}

	fmt.Println("All downloaded source code files deleted.")
}

func main() {
	linksFile := flag.String("links", "", "Filename containing the list of URLs")
	outputFile := flag.String("output", "", "Filename to save the filtered links")
	saveFile := flag.String("save", "", "Filename to save the source URL and its filtered links")
	flag.Parse()

	if *linksFile == "" {
		fmt.Println("Please provide the filename containing the list of URLs")
		flag.Usage()
		return
	}

	if *outputFile == "" {
		fmt.Println("Please provide the filename to save the filtered links")
		flag.Usage()
		return
	}

	file, err := os.Open(*linksFile)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", *linksFile, err)
		return
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		if url != "" {
			urls = append(urls, url)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file %s: %v\n", *linksFile, err)
		return
	}

	startTime := time.Now()

	var wg sync.WaitGroup
	ch := make(chan string, len(urls))
	outputChan := make(chan string, len(urls))
	saveChan := make(chan string, len(urls))

	go func() {
		outputFile, err := os.Create(*outputFile)
		if err != nil {
			fmt.Printf("Error creating output file %s: %v\n", *outputFile, err)
			close(outputChan)
			return
		}
		defer outputFile.Close()
		
		writer := bufio.NewWriter(outputFile)
		defer writer.Flush()

		for link := range outputChan {
			_, err := writer.WriteString(link + "\n")
			if err != nil {
				fmt.Printf("Error writing to output file %s: %v\n", *outputFile, err)
			}
		}
	}()

	if *saveFile != "" {
		go func() {
			saveFile, err := os.Create(*saveFile)
			if err != nil {
				fmt.Printf("Error creating save file %s: %v\n", *saveFile, err)
				close(saveChan)
				return
			}
			defer saveFile.Close()

			writer := bufio.NewWriter(saveFile)
			defer writer.Flush()

			for entry := range saveChan {
				_, err := writer.WriteString(entry + "\n")
				if err != nil {
					fmt.Printf("Error writing to save file %s: %v\n", *saveFile, err)
				}
			}
		}()
	}

	for _, url := range urls {
		wg.Add(1)
		go downloadSourceCode(url, &wg, ch)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for filename := range ch {
		searchAndHighlightLinks(filename, filterStrings, filename, outputChan, saveChan)
	}

	close(outputChan)
	if *saveFile != "" {
		close(saveChan)
	}

	endTime := time.Now()
	fmt.Printf("Execution time: %v seconds\n", endTime.Sub(startTime).Seconds())

	// Uncomment the line below to delete downloaded files after processing
	// deleteDownloadedFiles()
}
