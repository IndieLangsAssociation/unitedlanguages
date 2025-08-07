package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const RAW_BASE_URL = "https://raw.githubusercontent.com/IndieLangsAssociation/packages/main/"
const API_BASE_URL = "https://api.github.com/repos/IndieLangsAssociation/packages/contents"

type FileItem struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"`
	DownloadURL string `json:"download_url"`
}

func getFileText(gitPath string) string {
	url := RAW_BASE_URL + "/" + gitPath
	fmt.Printf("[INFO] Fetching raw text from: %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("[ERROR] getFileText failed for %s: %v\n", gitPath, err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("[ERROR] HTTP Status %d while getting %s\n", resp.StatusCode, gitPath)
		os.Exit(1)
	}
	data, _ := io.ReadAll(resp.Body)
	return string(data)
}

func getDirChildrenJson(filePath string) []FileItem {
	url := API_BASE_URL + "/" + filePath
	fmt.Printf("[INFO] Fetching directory listing from: %s\n", url)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		fmt.Printf("[ERROR] Failed to fetch dir %s: %v\n", filePath, err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	var items []FileItem
	err = json.NewDecoder(resp.Body).Decode(&items)
	if err != nil {
		fmt.Printf("[ERROR] Failed to parse JSON for dir %s: %v\n", filePath, err)
		os.Exit(1)
	}
	return items
}

func downloadBranched(gitPath, systemPath string) {
	fmt.Printf("[INFO] Creating directory (if not exists): %s\n", systemPath)
	os.MkdirAll(systemPath, 0755)
	items := getDirChildrenJson(gitPath)
	for _, item := range items {
		localPath := filepath.Join(systemPath, item.Name)
		fmt.Printf("[INFO] Processing item: %s (%s)\n", item.Path, item.Type)
		if item.Type == "file" {
			resp, err := http.Get(item.DownloadURL)
			if err != nil || resp.StatusCode != 200 {
				fmt.Printf("[ERROR] Failed to download file %s\n", item.Path)
				continue
			}
			data, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			os.WriteFile(localPath, data, 0644)
			fmt.Printf("[SUCCESS] Downloaded: %s -> %s\n", item.Path, localPath)
		} else if item.Type == "dir" {
			fmt.Printf("[INFO] Entering directory: %s\n", item.Path)
			downloadBranched(item.Path, localPath)
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("[USAGE] ulang install <package1> <package2> ...")
		os.Exit(1)
	}

	execPath, _ := os.Executable()
	basePath := filepath.Dir(execPath)
	configPath := filepath.Join(basePath, "config.json")
	fmt.Printf("[INFO] Loading config from: %s\n", configPath)
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("[FATAL] Failed to load config: %v\n", err)
		os.Exit(1)
	}

	var config struct {
		PackagesStorageDir string `json:"PackagesStorageDir"`
	}
	json.Unmarshal(configFile, &config)

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "install":
		fmt.Println("[INFO] Fetching registry metadata (data.json)")
		metaDataRaw := getFileText("packages/data.json")
		var metaData map[string]map[string]string
		json.Unmarshal([]byte(metaDataRaw), &metaData)

		for _, pkg := range args {
			path, ok := metaData["packages"][pkg]
			if !ok {
				fmt.Printf("[ERROR] Package '%s' is not found in registry\n", pkg)
				continue
			}
			packageDir := getDirChildrenJson(path)
			var metaFile *FileItem
			for _, item := range packageDir {
				if item.Name == "metadata.json" {
					metaFile = &item
					break
				}
			}
			if metaFile == nil {
				fmt.Printf("[ERROR] metadata.json not found in %s\n", pkg)
				continue
			}
			metaText := getFileText(metaFile.Path)
			var pkgInfo struct {
				Name     string            `json:"Name"`
				Latest   string            `json:"latest"`
				Versions map[string]string `json:"versions"`
			}
			json.Unmarshal([]byte(metaText), &pkgInfo)
			latestPath := pkgInfo.Versions[pkgInfo.Latest]
			fmt.Printf("[INFO] Downloading latest version: %s\n", pkgInfo.Latest)
			downloadBranched(latestPath, filepath.Join(config.PackagesStorageDir, pkgInfo.Name))
		}

	default:
		fmt.Printf("[ERROR] Unknown command: %s\n", cmd)
	}
}
