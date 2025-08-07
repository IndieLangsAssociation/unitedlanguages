package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

var (
	RAW_BASE_URL = "https://raw.githubusercontent.com/IndieLangsAssociation/unitedlanguages/main"
	API_BASE_URL = "https://api.github.com/repos/IndieLangsAssociation/unitedlanguages/contents"
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

func addToUserPATH(folder string) {
	key, err := syscall.UTF16PtrFromString("Environment")
	valueName, _ := syscall.UTF16PtrFromString("Path")

	hKey := syscall.Handle(0x80000001) // HKEY_CURRENT_USER
	var regKey syscall.Handle
	err = syscall.RegOpenKeyEx(hKey, key, 0, syscall.KEY_READ|syscall.KEY_WRITE, &regKey)
	if err != nil {
		// Create new key if it doesn't exist

		err = syscall.RegCreateKeyEx(hKey, key, 0, nil, 0, syscall.KEY_WRITE, nil, &regKey, nil)
		if err != nil {
			fmt.Printf("[ERROR] Could not create registry key: %v\n", err)
			return
		}
		defer syscall.RegCloseKey(regKey)
		syscall.RegSetValueEx(regKey, valueName, 0, syscall.REG_EXPAND_SZ, syscall.StringToUTF16Bytes(folder))
		fmt.Println("[INFO] Path created and set.")
		return
	}
	defer syscall.RegCloseKey(regKey)

	var buf [1 << 12]uint16
	var bufLen uint32 = uint32(len(buf))
	syscall.RegQueryValueEx(regKey, valueName, nil, nil, (*byte)(unsafe.Pointer(&buf[0])), &bufLen)
	curPath := syscall.UTF16ToString(buf[:])

	paths := strings.Split(curPath, ";")
	for _, p := range paths {
		if p == folder {
			fmt.Println("[INFO] Path already exists")
			return
		}
	}
	paths = append(paths, folder)
	newPath := strings.Join(paths, ";")
	syscall.RegSetValueEx(regKey, valueName, 0, syscall.REG_EXPAND_SZ, syscall.StringToUTF16Bytes(newPath))
	fmt.Println("[INFO] Successfully added to user PATH.")
}

func broadcastEnvChange() {
	HWND_BROADCAST := uintptr(0xffff)
	WM_SETTINGCHANGE := uintptr(0x001A)
	SMTO_ABORTIFHUNG := uintptr(0x0002)
	user32 := syscall.NewLazyDLL("user32.dll")
	proc := user32.NewProc("SendMessageTimeoutW")
	proc.Call(HWND_BROADCAST, WM_SETTINGCHANGE, 0, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("Environment"))), SMTO_ABORTIFHUNG, 5000, 0)
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

	addToUserPATH(installDir)
	broadcastEnvChange()

	fmt.Printf("[LOG] ADDED %s to user PATH\n", installDir)
	fmt.Printf("[LOG] Installed: %s; edit config.json to your needs.\n", strings.Join(files, ", "))
	fmt.Println("Press Enter to exit...")
	exec.Command("cmd", "/c", "pause").Run()
}
