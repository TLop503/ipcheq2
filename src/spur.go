package src

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
)

func checkSpur(ip string) (string, error) {
	url := "https://spur.us/context/" + ip
	// Create a request so we can attach the session cookie header
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	// Attach mandatory Spur session cookie
	if SpurCookie == "" {
		log.Printf("spur session cookie not initialized")
		return "", errors.New("spur session cookie not initialized")
	}
	req.Header.Set("Cookie", SpurCookie)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	content, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	/*
		parse out meta description from the HTML
		meta_pattern = r'<meta\s+name=["\']description["\']\s+content=["\']([^"\']+)["\']'
		match = re.search(meta_pattern, html_content, re.IGNORECASE)
	*/
	pattern := `<meta\s+name=["']description["']\s+content=["']([^"']+)["']`
	re := regexp.MustCompile(pattern)
	fmt.Println(string(content))

	match := re.FindStringSubmatch(string(content))
	if len(match) > 0 {
		// truncate for readability first
		if i := strings.Index(match[1], "VPN."); i != -1 {
			return match[1][:i+len("VPN.")], nil
		} else if i := strings.Index(match[1], "activity."); i != -1 {
			return match[1][:i+len("activity.")], nil
		} else {
			return match[1], nil
		}
	}

	return "Spur query failed!", nil
}
