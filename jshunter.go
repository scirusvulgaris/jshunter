package main

import (
    "bufio"
    "flag"
    "fmt"
    "io/ioutil"
    "net/http"
    "net/url"
    "os"
    "regexp"
    "strings"
    "sync"
    "time"
)


var colors = map[string]string{
    "RED":    "\033[0;31m",
    "GREEN":  "\033[0;32m",
    "BLUE":   "\033[0;34m",
    "YELLOW": "\033[0;33m",
    "CYAN":   "\033[0;36m",
    "PURPLE": "\033[0;35m",
    "NC":     "\033[0m",
}



var (
    //regex patterns
    regexPatterns = map[string]*regexp.Regexp{
	"google_api":                    regexp.MustCompile(`AIza[0-9A-Za-z-_]{35}`),
	"firebase":                      regexp.MustCompile(`AAAA[A-Za-z0-9_-]{7}:[A-Za-z0-9_-]{140}`),
	"google_captcha":                regexp.MustCompile(`6L[0-9A-Za-z-_]{38}|^6[0-9a-zA-Z_-]{39}$`),
	"google_oauth":                  regexp.MustCompile(`ya29\.[0-9A-Za-z\-_]+`),
	"amazon_aws_access_key_id":      regexp.MustCompile(`A[SK]IA[0-9A-Z]{16}`),
	"amazon_mws_auth_token":         regexp.MustCompile(`amzn\\.mws\\.[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`),
	"amazon_aws_url":                regexp.MustCompile(`s3\.amazonaws.com[/]+|[a-zA-Z0-9_-]*\.s3\.amazonaws.com`),
	"amazon_aws_url2":               regexp.MustCompile(`([a-zA-Z0-9-._]+\.s3\.amazonaws\.com|s3://[a-zA-Z0-9-._]+|s3-[a-zA-Z0-9-._/]+|s3.amazonaws.com/[a-zA-Z0-9-._]+|s3.console.aws.amazon.com/s3/buckets/[a-zA-Z0-9-._]+)`),
	"facebook_access_token":         regexp.MustCompile(`EAACEdEose0cBA[0-9A-Za-z]+`),
	"authorization_basic":           regexp.MustCompile(`basic [a-zA-Z0-9=:_\+\/-]{5,100}`),
	"authorization_bearer":          regexp.MustCompile(`bearer [a-zA-Z0-9_\-\.=:_\+\/]{5,100}`),
	"authorization_api":             regexp.MustCompile(`api[key|_key|\s+]+[a-zA-Z0-9_\-]{5,100}`),
	"mailgun_api_key":               regexp.MustCompile(`key-[0-9a-zA-Z]{32}`),
	"twilio_api_key":                regexp.MustCompile(`SK[0-9a-fA-F]{32}`),
	"twilio_account_sid":            regexp.MustCompile(`AC[a-zA-Z0-9_\-]{32}`),
	"twilio_app_sid":                regexp.MustCompile(`AP[a-zA-Z0-9_\-]{32}`),
	"paypal_braintree_access_token": regexp.MustCompile(`access_token\$production\$[0-9a-z]{16}\$[0-9a-f]{32}`),
	"square_oauth_secret":           regexp.MustCompile(`sq0csp-[0-9A-Za-z\-_]{43}|sq0[a-z]{3}-[0-9A-Za-z\-_]{22,43}`),
	"square_access_token":           regexp.MustCompile(`sqOatp-[0-9A-Za-z\-_]{22}|EAAA[a-zA-Z0-9]{60}`),
	"stripe_standard_api":           regexp.MustCompile(`sk_live_[0-9a-zA-Z]{24}`),
	"stripe_restricted_api":         regexp.MustCompile(`rk_live_[0-9a-zA-Z]{24}`),
	"github_access_token":           regexp.MustCompile(`[a-zA-Z0-9_-]*:[a-zA-Z0-9_\-]+@github\.com*`),
	"rsa_private_key":               regexp.MustCompile(`-----BEGIN RSA PRIVATE KEY-----`),
	"ssh_dsa_private_key":           regexp.MustCompile(`-----BEGIN DSA PRIVATE KEY-----`),
	"ssh_dc_private_key":            regexp.MustCompile(`-----BEGIN EC PRIVATE KEY-----`),
	"pgp_private_block":             regexp.MustCompile(`-----BEGIN PGP PRIVATE KEY BLOCK-----`),
	"json_web_token":                regexp.MustCompile(`ey[A-Za-z0-9-_=]+\.[A-Za-z0-9-_=]+\.?[A-Za-z0-9-_.+/=]*$`),
	"slack_token":                   regexp.MustCompile(`\"api_token\":\"(xox[a-zA-Z]-[a-zA-Z0-9-]+)\"`),
	"SSH_privKey":                   regexp.MustCompile(`([-]+BEGIN [^\s]+ PRIVATE KEY[-]+[\s]*[^-]*[-]+END [^\s]+ PRIVATE KEY[-]+)`),
	"Heroku API KEY":                regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`),
	"slack_webhook_url":             regexp.MustCompile(`https://hooks.slack.com/services/[A-Za-z0-9]+/[A-Za-z0-9]+/[A-Za-z0-9]+`),
	"heroku_api_key":                regexp.MustCompile(`[hH]eroku[a-zA-Z0-9]{32}`),
	"dropbox_access_token":          regexp.MustCompile(`sl\.[A-Za-z0-9_-]+`),
	"salesforce_access_token":       regexp.MustCompile(`00D[0-9A-Za-z]{15,18}![A-Za-z0-9]{40}`),
	"twitter_bearer_token":          regexp.MustCompile(`AAAAAAAAAAAAAAAAAAAAA[A-Za-z0-9%]{30,45}`),
	"firebase_url":                  regexp.MustCompile(`https://[a-z0-9-]+\.firebaseio\.com`),
	"pem_private_key":               regexp.MustCompile(`-----BEGIN (?:[A-Z ]+ )?PRIVATE KEY-----`),
	"google_cloud_sa_key":           regexp.MustCompile(`"type": "service_account"`),
	"stripe_publishable_key":        regexp.MustCompile(`pk_live_[0-9a-zA-Z]{24}`),
	"azure_storage_account_key":     regexp.MustCompile(`[A-Za-z0-9+/=]{88}`),
	"instagram_access_token":        regexp.MustCompile(`IGQV[A-Za-z0-9._-]{10,}`),
	"stripe_test_publishable_key":   regexp.MustCompile(`pk_test_[0-9a-zA-Z]{24}`),
	"stripe_test_secret_key":        regexp.MustCompile(`sk_test_[0-9a-zA-Z]{24}`),
	"slack_bot_token":               regexp.MustCompile(`xoxb-[A-Za-z0-9-]{24,34}`),
	"slack_user_token":              regexp.MustCompile(`xoxp-[A-Za-z0-9-]{24,34}`),
	"slack_webhook":                 regexp.MustCompile(`https://hooks.slack.com/services/T[a-zA-Z0-9_]+/B[a-zA-Z0-9_]+/[a-zA-Z0-9_]+`),
	"discord_webhook_url":           regexp.MustCompile(`https://discord(?:app)?\.com/api/webhooks/[0-9]{18,20}/[A-Za-z0-9_-]{64,}`),
	"discord_bot_token":             regexp.MustCompile(`[MN][A-Za-z\d]{23}\.[\w-]{6}\.[\w-]{27}`),
	"okta_api_token":                regexp.MustCompile(`00[a-zA-Z0-9]{30}\.[a-zA-Z0-9\-_]{30,}\.[a-zA-Z0-9\-_]{30,}`),
	"sendgrid_api_key":              regexp.MustCompile(`SG\.[A-Za-z0-9_-]{22}\.[A-Za-z0-9_-]{43}`),
	"mapbox_access_token":           regexp.MustCompile(`pk\.[a-zA-Z0-9]{60}\.[a-zA-Z0-9]{22}`),
	"gitlab_personal_access_token":  regexp.MustCompile(`glpat-[A-Za-z0-9\-]{20}`),
	"datadog_api_key":               regexp.MustCompile(`ddapi_[a-zA-Z0-9]{32}`),
	"shopify_access_token":          regexp.MustCompile(`shpat_[A-Za-z0-9]{32}`),
	"atlassian_access_token":        regexp.MustCompile(`[0-9a-z]{3}\.[0-9a-z]{1}\.[0-9a-z]{3}`),
	"crowdstrike_api_key":           regexp.MustCompile(`[\w-]{32}\.[\w-]{16}`),
	"quickbooks_api_key":            regexp.MustCompile(`A[0-9a-f]{32}`),
	"cisco_api_key":                 regexp.MustCompile(`cisco[A-Za-z0-9]{30}`),
	"cisco_access_token":            regexp.MustCompile(`access_token=\w+`),
	"segment_write_key":             regexp.MustCompile(`sk_[A-Za-z0-9]{32}`),
	"tiktok_access_token":           regexp.MustCompile(`tiktok_access_token=[a-zA-Z0-9_]+`),
	"slack_client_secret":           regexp.MustCompile(`xoxs-[0-9]{1,9}.[0-9A-Za-z]{1,12}.[0-9A-Za-z]{24,64}`),
        "phone_number":                  regexp.MustCompile(`^\+\d{9,14}$`),
        "email":                         regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
    }

    asciiArt = `
         ________             __         
     __ / / __/ /  __ _____  / /____ ____
    / // /\ \/ _ \/ // / _ \/ __/ -_) __/
    \___/___/_//_/\_,_/_//_/\__/\__/_/  

     v0.1                         Created by cc1a2b
    `
)

func main() {
    // Define command-line
    var (
        url, list, jsFile, output, regex, cookies, proxy string
        threads                                           int
        quiet, help                                       bool
    )


    flag.StringVar(&url, "u", "", "Input a URL")
    flag.StringVar(&url, "url", "", "Input a URL")
    flag.StringVar(&list, "l", "", "Input a file with URLs (.txt)")
    flag.StringVar(&list, "list", "", "Input a file with URLs (.txt)")
    flag.StringVar(&jsFile, "f", "", "Path to JavaScript file")
    flag.StringVar(&jsFile, "file", "", "Path to JavaScript file")
    flag.StringVar(&output, "o", "output.txt", "Output file path (default: output.txt)")
    flag.StringVar(&output, "output", "output.txt", "Output file path (default: output.txt)")
    flag.StringVar(&regex, "r", "", "RegEx for filtering endpoints")
    flag.StringVar(&regex, "regex", "", "RegEx for filtering endpoints")
    flag.StringVar(&cookies, "c", "", "Cookies for authenticated JS files")
    flag.StringVar(&cookies, "cookies", "", "Cookies for authenticated JS files")
    flag.StringVar(&proxy, "p", "", "Set proxy (host:port)")
    flag.StringVar(&proxy, "proxy", "", "Set proxy (host:port)")
    flag.IntVar(&threads, "t", 5, "Number of concurrent threads")
    flag.IntVar(&threads, "threads", 5, "Number of concurrent threads")
    flag.BoolVar(&quiet, "q", false, "Quiet mode: suppress ASCII art output")
    flag.BoolVar(&quiet, "quiet", false, "Quiet mode: suppress ASCII art output")
    flag.BoolVar(&help, "h", false, "Display help message")
    flag.BoolVar(&help, "help", false, "Display help message")


    flag.Parse()

    if help {
        customHelp()
        return
    }


    if len(os.Args) == 1 || quiet {
        time.Sleep(100 * time.Millisecond)
        customHelp()
        return
    }


    if url == "" && list == "" && jsFile == "" {
        if isInputFromStdin() {
            scanner := bufio.NewScanner(os.Stdin)
            for scanner.Scan() {
                inputURL := scanner.Text()
                // Process each input URL
                processInputs(inputURL, "", output, regex, cookies, proxy, threads)
            }
            if err := scanner.Err(); err != nil {
                fmt.Fprintln(os.Stderr, "Error reading from stdin:", err)
            }
            return
        }
        fmt.Println("Error: Either -u, -l, or -f must be provided.")
        os.Exit(1)
    }


    if !quiet {
        time.Sleep(100 * time.Millisecond)
        fmt.Println(asciiArt)
    }


    if quiet {
        disableColors()
    }


    if jsFile != "" {
        processJSFile(jsFile, regex)
    }


    processInputs(url, list, output, regex, cookies, proxy, threads)
}


func customHelp() {
    fmt.Println(asciiArt)
    fmt.Println("Usage:")
    fmt.Println("  -u, --url URL                 Input a URL")
    fmt.Println("  -l, --list FILE.txt           Input a file with URLs (.txt)")
    fmt.Println("  -f, --file FILE.js            Path to JavaScript file")
    fmt.Println()
    fmt.Println("Options:")
    fmt.Println("  -t, --threads INT             Number of concurrent threads (default: 5)")
    fmt.Println("  -c, --cookies <cookies>       Cookies for authenticated JS files")
    fmt.Println("  -p, --proxy host:port         Set proxy (host:port), e.g., 127.0.0.1:8080 for Burp Suite")
    fmt.Println("  -nc, --no-color               Disable color output")
    fmt.Println("  -q, --quiet                   Suppress ASCII art output")
    fmt.Println("  -o, --output FILENAME.txt     Output file path (default: output.txt)")
    fmt.Println("  -r, --regex <pattern>         RegEx for filtering endpoints")
    fmt.Println("  -h, --help                    Display this help message")
}

func processStdin(output, regex, cookies, proxy string, threads int) {
    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan() {
        line := scanner.Text()
        fmt.Println("Processing line from stdin:", line)

    }
    if err := scanner.Err(); err != nil {
        fmt.Fprintln(os.Stderr, "Error reading from stdin:", err)
    }
}


func isInputFromStdin() bool {
    fi, err := os.Stdin.Stat()
    if err != nil {
        fmt.Println("Error checking stdin:", err)
        return false
    }
    return fi.Mode()&os.ModeCharDevice == 0
}


func disableColors() {
    for k := range colors {
        colors[k] = ""
    }
}


func processJSFile(jsFile, regex string) {
    if _, err := os.Stat(jsFile); os.IsNotExist(err) {
        fmt.Printf("[%sERROR%s] File not found: %s\n", colors["RED"], colors["NC"], jsFile)
    } else if err != nil {
        fmt.Printf("[%sERROR%s] Unable to access file %s: %v\n", colors["RED"], colors["NC"], jsFile, err)
    } else {
        fmt.Printf("[%sFOUND%s] FILE: %s\n", colors["RED"], colors["NC"], jsFile)
        searchForSensitiveData(jsFile, regex, "", "")
    }
}


func processInputs(url, list, output, regex, cookie, proxy string, threads int) {
    var wg sync.WaitGroup
    urlChannel := make(chan string)

    var fileWriter *os.File
    if output != "" {
        var err error
        fileWriter, err = os.Create(output)
        if err != nil {
            fmt.Printf("Error creating output file: %v\n", err)
            return
        }
        defer fileWriter.Close()
    }

    for i := 0; i < threads; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for u := range urlChannel {
                _, sensitiveData := searchForSensitiveData(u, regex, cookie, proxy)

                if fileWriter != nil {
                    fmt.Fprintln(fileWriter, "URL:", u)
                    for name, matches := range sensitiveData {
                        for _, match := range matches {
                            fmt.Fprintf(fileWriter, "Sensitive Data [%s]: %s\n", name, match)
                        }
                    }
                } else {
                    fmt.Println("URL:", u)
                    for name, matches := range sensitiveData {
                        for _, match := range matches {
                            fmt.Printf("Sensitive Data [%s]: %s\n", name, match)
                        }
                    }
                }
            }
        }()
    }

    if err := enqueueURLs(url, list, urlChannel, regex); err != nil {
        fmt.Printf("Error in input processing: %v\n", err)
        close(urlChannel)
        return
    }

    close(urlChannel)
    wg.Wait()
}


func enqueueURLs(url, list string, urlChannel chan<- string, regex string) error {
    if list != "" {
        return enqueueFromFile(list, urlChannel)
    } else if url != "" {
        enqueueSingleURL(url, urlChannel, regex)
    } else {
        enqueueFromStdin(urlChannel)
    }
    return nil
}

func enqueueFromFile(filename string, urlChannel chan<- string) error {
    file, err := os.Open(filename)
    if err != nil {
        return fmt.Errorf("Error opening file: %w", err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        urlChannel <- scanner.Text()
    }
    return scanner.Err()
}

func enqueueSingleURL(url string, urlChannel chan<- string, regex string) {
    if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
        urlChannel <- url
    } else {
        processJSFile(url, regex)
    }
}

func enqueueFromStdin(urlChannel chan<- string) {
    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan() {
        urlChannel <- scanner.Text()
    }
    if err := scanner.Err(); err != nil {
        fmt.Printf("Error reading from stdin: %v\n", err)
    }
}


func searchForSensitiveData(urlStr, regex, cookie, proxy string) (string, map[string][]string) {
    var client *http.Client

    if proxy != "" {
        proxyURL, err := url.Parse(proxy)
        if err != nil {
            fmt.Printf("Invalid proxy URL: %v\n", err)
            return urlStr, nil
        }
        client = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)}}
    } else {
        client = &http.Client{}
    }

    var sensitiveData map[string][]string

    if strings.HasPrefix(urlStr, "http://") || strings.HasPrefix(urlStr, "https://") {
        req, err := http.NewRequest("GET", urlStr, nil)
        if err != nil {
            fmt.Printf("Failed to create request for URL %s: %v\n", urlStr, err)
            return urlStr, nil
        }

        if cookie != "" {
            req.Header.Set("Cookie", cookie)
        }

        resp, err := client.Do(req)
        if err != nil {
            fmt.Printf("Failed to fetch URL %s: %v\n", urlStr, err)
            return urlStr, nil
        }
        defer resp.Body.Close()

        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            fmt.Printf("Error reading response body: %v\n", err)
            return urlStr, nil
        }

        sensitiveData = reportMatches(urlStr, body, regexPatterns)
    } else {
        body, err := ioutil.ReadFile(urlStr)
        if err != nil {
            fmt.Printf("Error reading local file %s: %v\n", urlStr, err)
            return urlStr, nil
        }

        sensitiveData = reportMatches(urlStr, body, regexPatterns)
    }

    return urlStr, sensitiveData
}


func reportMatches(source string, body []byte, regexPatterns map[string]*regexp.Regexp) map[string][]string {
    matchesMap := make(map[string][]string)

    for name, pattern := range regexPatterns {
        if pattern.Match(body) {
            matches := pattern.FindAllString(string(body), -1)
            if len(matches) > 0 {
                matchesMap[name] = append(matchesMap[name], matches...)
            }
        }
    }

    if len(matchesMap) > 0 {
        fmt.Printf("[%sFOUND%s] Sensitive data at: %s\n", colors["RED"], colors["NC"], source)
        for name, matches := range matchesMap {
            for _, match := range matches {
                fmt.Printf("%s ==>> %s\n", name, match)
            }
        }
    } else {
        fmt.Printf("[%sMISSING%s] No sensitive data found at: %s\n", colors["BLUE"], colors["NC"], source)
    }

    return matchesMap
}
