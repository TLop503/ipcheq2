package src

import (
	"errors"
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
	// First, try to extract the human-friendly paragraph that contains the summary
	// e.g. <p data-slot="text">192.145.119.7 belongs to the Nord VPN anonymization network. ...</p>
	pRe := regexp.MustCompile(`(?si)<p[^>]*data-slot=["']text["'][^>]*>(.*?)</p>`)
	if pm := pRe.FindStringSubmatch(string(content)); len(pm) > 1 {
		// strip any HTML tags inside the paragraph
		stripRe := regexp.MustCompile(`<[^>]+>`)
		text := stripRe.ReplaceAllString(pm[1], "")
		text = strings.TrimSpace(text)
		if text != "" {
			return text, nil
		}
	}

	// fallback: parse meta description (older behavior)
	pattern := `<meta\s+name=["']description["']\s+content=["']([^"']+)["']`
	re := regexp.MustCompile(pattern)

	match := re.FindStringSubmatch(string(content))
	if len(match) > 0 {
		meta := match[1]
		// Try to find a provider name that ends with 'VPN', e.g. 'Nord VPN'
		providerRe := regexp.MustCompile(`(?i)([A-Za-z0-9&\-_ ]+VPN)\b`)
		if p := providerRe.FindStringSubmatch(meta); len(p) > 1 {
			prov := strings.TrimSpace(p[1])
			return "VPN: " + prov, nil
		}

		// Explicit negative indicators
		notRe := regexp.MustCompile(`(?i)\b(not (a )?vpn|no vpn|not vpn)\b`)
		if notRe.MatchString(meta) {
			return "Not VPN", nil
		}

		// If the meta mentions VPN but no provider parsed, return generic VPN
		if strings.Contains(strings.ToLower(meta), "vpn") {
			return "VPN", nil
		}

		// fallback to previous activity truncation
		if i := strings.Index(meta, "activity."); i != -1 {
			return meta[:i+len("activity.")], nil
		}

		return meta, nil
	}

	return "Spur query failed!", nil
}
