package virustotal

import "github.com/VirusTotal/vt-go"

func getInt(result *vt.Object, path string) int {
	if v, err := result.GetInt64(path); err == nil {
		return int(v)
	}
	return 0
}

func getString(result *vt.Object, path string) string {
	if v, err := result.GetString(path); err == nil {
		return v
	}
	return ""
}

func getStringSlice(result *vt.Object, path string) []string {
	if v, err := result.GetStringSlice(path); err == nil {
		return v
	}
	return nil
}
