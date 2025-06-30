package src

import (
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
)

func checkSpur(ip string) (string, error) {
	url := "https://spur.us/context/" + ip

	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	content, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	/*
		parse out meta description from the HTML
		meta_pattern = r'<meta\s+name=["\']description["\']\s+content=["\']([^"\']+)["\']'
		match = re.search(meta_pattern, html_content, re.IGNORECASE)
	*/
	pattern := `<meta\s+name=["']description["']\s+content=["']([^"']+)["']`
	re := regexp.MustCompile(pattern)

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
