package main

import (
    "bufio"
    "crypto/tls"
    "encoding/json"
    "flag"
    "io/ioutil"
    "log"
    "math/rand"
    "net/http"
    "os"
    "strings"
    "sync"
    "time"
)

// ANSI escape sequences for colors
const (
    GREEN = "\033[92m"
    RED   = "\033[91m"
    BOLD  = "\033[1m"
    RESET = "\033[0m"
)

// Global counters for statistics
var (
    bugBountyFoundCount    int
    bugBountyNotFoundCount int
    foundBountyURLs        []string
    mu                     sync.Mutex
)

// List of common user-agent strings
var userAgents = []string{
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0.3 Safari/605.1.15",
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0",
    "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36",
}

func getRandomUserAgent() string {
    return userAgents[rand.Intn(len(userAgents))]
}

func fetchURL(url string, wg *sync.WaitGroup) {
    defer wg.Done()

    // Endpoints to check
    endpoints := []string{
        "/bugbountytesting.txt",
        "/upload/bugbountytesting.txt",
        "/bugbountytesting.json",
    }

    // Custom HTTP client with disabled SSL verification
    tr := &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    }
    client := &http.Client{
        Timeout:   10 * time.Second,
        Transport: tr,
    }

    for _, endpoint := range endpoints {
        putURL := url + endpoint
        userAgent := getRandomUserAgent()
        var putData string
        var contentType string

        if strings.HasSuffix(endpoint, ".json") {
            data := map[string]string{"message": "bugbountytestingxyz"}
            jsonData, _ := json.Marshal(data)
            putData = string(jsonData)
            contentType = "application/json"
        } else {
            putData = "bugbountytestingxyz"
            contentType = "text/plain"
        }

        // PUT request
        req, err := http.NewRequest("PUT", putURL, strings.NewReader(putData))
        if err != nil {
            log.Printf("Error creating PUT request: %v\n", err)
            continue
        }
        req.Header.Set("Content-Type", contentType)
        req.Header.Set("User-Agent", userAgent)

        resp, err := client.Do(req)
        if err != nil {
            log.Printf("Error making PUT request to %s: %v\n", putURL, err)
            continue
        }
        resp.Body.Close()

        // GET request
        req, err = http.NewRequest("GET", putURL, nil)
        if err != nil {
            log.Printf("Error creating GET request: %v\n", err)
            continue
        }
        req.Header.Set("User-Agent", userAgent)

        resp, err = client.Do(req)
        if err != nil {
            log.Printf("Error making GET request to %s: %v\n", putURL, err)
            continue
        }
        body, err := ioutil.ReadAll(resp.Body)
        resp.Body.Close()
        if err != nil {
            log.Printf("Error reading response body: %v\n", err)
            continue
        }

        // Check if "bugbountytestingxyz" is found in the response
        if strings.Contains(string(body), "bugbountytestingxyz") {
            mu.Lock()
            bugBountyFoundCount++
            foundBountyURLs = append(foundBountyURLs, putURL)
            mu.Unlock()
            log.Printf("URL: %s\nEndpoint: %s\n%sUploaded File found in the response%s\n%s\n",
                url, endpoint, GREEN, RESET, strings.Repeat("=", 60))
        } else {
            mu.Lock()
            bugBountyNotFoundCount++
            mu.Unlock()
            log.Printf("URL: %s\nEndpoint: %s\n%sUploaded File not found in the response%s\n%s\n",
                url, endpoint, RED, RESET, strings.Repeat("=", 60))
        }
    }
}

func processURLs(filePath string, batchSize int) {
    file, err := os.Open(filePath)
    if err != nil {
        log.Fatalf("Error opening file: %v\n", err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    var wg sync.WaitGroup

    for scanner.Scan() {
        url := scanner.Text()
        if url == "" {
            continue
        }
        wg.Add(1)
        go fetchURL(url, &wg)
        time.Sleep(10 * time.Millisecond) // Throttle requests to avoid server overload
    }
    wg.Wait()

    if err := scanner.Err(); err != nil {
        log.Fatalf("Error reading file: %v\n", err)
    }
}

func saveResults(outputPath string) {
    resultFile, err := os.Create(outputPath)
    if err != nil {
        log.Fatalf("Error creating result file: %v\n", err)
    }
    defer resultFile.Close()

    for _, foundURL := range foundBountyURLs {
        resultFile.WriteString(foundURL + "\n")
    }
}

func main() {
    rand.Seed(time.Now().UnixNano())

    filePath := flag.String("file", "", "Path to the file containing URLs")
    outputPath := flag.String("output", "results.txt", "Path to the output file for results")
    flag.Parse()

    if *filePath == "" {
        log.Fatal("Please provide a file path using the -file flag.")
    }

    processURLs(*filePath, 100)

    // Print Bug Bounty found and not found counts
    log.Printf("\n%sRequest Analysis:%s\n", BOLD, RESET)
    log.Printf("File Upload Successful:   %s%s%d%s\n", GREEN, BOLD, bugBountyFoundCount, RESET)
    log.Printf("File Upload Unsuccessful:   %s%s%d%s\n", RED, BOLD, bugBountyNotFoundCount, RESET)

    // Save URLs with endpoints where Bug Bounty was found
    if len(foundBountyURLs) > 0 {
        saveResults(*outputPath)
    }
}
