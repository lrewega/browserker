package browser

import "runtime"

// FindChrome on the FS
func FindChrome() (string, string) {
	switch runtime.GOOS {
	case "windows":
		return "C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe", "C:\\Temp\\gcd\\"
	case "darwin":
		return "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome", "/tmp/gcd/"
	case "linux":
		return "/usr/bin/chromium-browser", "/tmp/gcd/"
	}
	return "", "tmp"
}

// FindKill based on OS
func FindKill(browser string) []string {
	switch runtime.GOOS {
	case "windows":
		return []string{"taskkill", "/F", "/IM", browser + ".exe"}
	case "darwin":
		return []string{"killall", "Google Chrome"}
	case "linux":
		return []string{"killall", browser}
	}
	return []string{""}
}
