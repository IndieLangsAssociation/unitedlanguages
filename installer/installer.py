import os, requests, winreg, ctypes

RAW_BASE_URL = 'https://raw.githubusercontent.com/IndieLangsAssociation/unitedlanguages/main'
API_BASE_URL = 'https://api.github.com/repos/IndieLangsAssociation/unitedlanguages/contents'

def getFileText(filePath):
    url = f"{RAW_BASE_URL}/{filePath}"
    print(f"[INFO] Fetching raw text from: {url}")
    try:
        r = requests.get(url)
        if r.status_code != 200:
            raise Exception(f"Failed to fetch text — Status code: {r.status_code}")
        return r.text
    except Exception as e:
        print(f"[ERROR] getFileText failed for {filePath}: {e}")
        raise

def addToPATH(folder_path):
    try:
        key = winreg.OpenKey(
            winreg.HKEY_CURRENT_USER,
            r"Environment",
            0,
            winreg.KEY_READ | winreg.KEY_WRITE
        )
        current_path, _ = winreg.QueryValueEx(key, "Path")
        paths = current_path.split(";") if current_path else []

        if folder_path not in paths:
            paths.append(folder_path)
            new_path = ";".join(paths)
            winreg.SetValueEx(key, "Path", 0, winreg.REG_EXPAND_SZ, new_path)
            print(f"[INFO] Successfully added to user PATH: {folder_path}")
        else:
            print(f"[INFO] Path already exists: {folder_path}")

        winreg.CloseKey(key)
    except FileNotFoundError:
        key = winreg.CreateKey(winreg.HKEY_CURRENT_USER, r"Environment")
        winreg.SetValueEx(key, "Path", 0, winreg.REG_EXPAND_SZ, folder_path)
        winreg.CloseKey(key)
        print(f"[INFO] Path created and set: {folder_path}")
    except Exception as e:
        print(f"[ERROR] Failed to set user PATH: {e}")

def broadcastEnvChange():
    HWND_BROADCAST = 0xFFFF
    WM_SETTINGCHANGE = 0x001A
    SMTO_ABORTIFHUNG = 0x0002
    ctypes.windll.user32.SendMessageTimeoutW(
        HWND_BROADCAST,
        WM_SETTINGCHANGE,
        0,
        "Environment",
        SMTO_ABORTIFHUNG,
        5000,
        None
    )


install_dir = os.path.join(os.path.expanduser("~"), "ulang")
os.makedirs(install_dir, exist_ok=True)

files = ["ulang.bat", "ulang.py", "config.json"]
for fname in files:
    path = os.path.join(install_dir, fname)
    with open(path, "w", encoding="utf-8") as f:
        f.write(getFileText(f"src/{fname}"))

addToPATH(install_dir)
broadcastEnvChange()

print(f"[LOG] ADDED {install_dir} to user PATH")
print(f"[LOG] Installed: {', '.join(files)}; edit config.json to your needs.")
input("Press Enter to exit...")
