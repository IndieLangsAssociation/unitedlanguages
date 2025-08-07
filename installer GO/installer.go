package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	RAW_BASE_URL = "https://raw.githubusercontent.com/IndieLangsAssociation/unitedlanguages/main"
)

func getFileText(filePath string) string {
	url := RAW_BASE_URL + "/" + filePath
	fmt.Printf("[INFO] Fetching raw text from: %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("[ERROR] Failed to fetch: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("[ERROR] Bad status code: %d\n", resp.StatusCode)
		os.Exit(1)
	}
	data, _ := io.ReadAll(resp.Body)
	return string(data)
}

func downloadBinary(destDir string) {
	var platform string
	switch runtime.GOOS {
	case "windows":
		platform = "win"
	case "linux":
		platform = "linux"
	case "darwin":
		platform = "mac"
	default:
		fmt.Printf("[ERROR] Unsupported OS: %s\n", runtime.GOOS)
		os.Exit(1)
	}

	url := fmt.Sprintf("%s/build/%s/ulang.exe", RAW_BASE_URL, platform)
	outPath := filepath.Join(destDir, "ulang.exe")

	fmt.Printf("[INFO] Downloading binary from: %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("[ERROR] Download failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("[ERROR] Binary not found for platform: %s\n", platform)
		os.Exit(1)
	}

	out, err := os.Create(outPath)
	if err != nil {
		fmt.Printf("[ERROR] Could not write file: %v\n", err)
		os.Exit(1)
	}
	defer out.Close()

	io.Copy(out, resp.Body)
	fmt.Printf("[INFO] Downloaded binary to: %s\n", outPath)
	os.Chmod(outPath, 0755)
}

func addToPathUnix(folder string) {
	shellProfile := filepath.Join(os.Getenv("HOME"), ".profile")
	if _, err := os.Stat(shellProfile); os.IsNotExist(err) {
		shellProfile = filepath.Join(os.Getenv("HOME"), ".bashrc")
	}
	exportLine := fmt.Sprintf("\nexport PATH=\"$PATH:%s\"\n", folder)

	content, _ := os.ReadFile(shellProfile)
	if strings.Contains(string(content), folder) {
		fmt.Println("[INFO] Path already exists in shell profile")
		return
	}

	f, err := os.OpenFile(shellProfile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("[WARN] Could not update shell profile: %v\n", err)
		return
	}
	defer f.Close()
	f.WriteString(exportLine)
	fmt.Printf("[INFO] Added to shell profile: %s\n", shellProfile)
}

func addToPathWindows(folder string) {
	cmd := exec.Command("powershell", "-Command", fmt.Sprintf(`[Environment]::SetEnvironmentVariable("Path", "$Env:Path + ";%s"", "User")`, folder))
	err := cmd.Run()
	if err != nil {
		fmt.Printf("[ERROR] Failed to set PATH: %v\n", err)
		return
	}
	fmt.Println("[INFO] Added to user PATH via PowerShell")
}

func main() {
	homeDir, _ := os.UserHomeDir()
	installDir := filepath.Join(homeDir, "ulang")
	fmt.Printf("[INFO] Creating install dir: %s\n", installDir)
	os.MkdirAll(installDir, 0755)

	files := []string{"ulang.bat", "ulang.py", "config.json"}
	for _, name := range files {
		text := getFileText("src/" + name)
		path := filepath.Join(installDir, name)
		os.WriteFile(path, []byte(text), 0644)
		fmt.Printf("[INFO] Installed: %s\n", path)
	}

	downloadBinary(installDir)

	switch runtime.GOOS {
	case "windows":
		addToPathWindows(installDir)
	case "linux", "darwin":
		addToPathUnix(installDir)
	default:
		fmt.Printf("[ERROR] Unsupported OS: %s\n", runtime.GOOS)
	}

	fmt.Printf("[LOG] ADDED %s to user PATH\n", installDir)
	fmt.Printf("[LOG] Installed: %s; edit config.json to your needs.\n", strings.Join(files, ", "))
	fmt.Println("Press Enter to exit...")
	fmt.Scanln()
}
