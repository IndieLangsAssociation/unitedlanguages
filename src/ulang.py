import json
import sys, os, requests

RAW_BASE_URL = 'https://raw.githubusercontent.com/IndieLangsAssociation/packages/main'
API_BASE_URL = 'https://api.github.com/repos/IndieLangsAssociation/packages/contents'

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

def downloadBranched(gitPath, systemPath):
    print(f"[INFO] Creating directory (if not exists): {systemPath}")
    os.makedirs(systemPath, exist_ok=True)

    print(f"[INFO] Getting children of Git path: {gitPath}")
    try:
        items = getDirChildrenJson(gitPath)
    except Exception as e:
        print(f"[ERROR] Cannot access directory {gitPath}: {e}")
        return

    for item in items:
        localPath = os.path.join(systemPath, item["name"])
        print(f"[INFO] Processing item: {item['path']} ({item['type']})")
        if item["type"] == "file":
            try:
                print(f"[INFO] Downloading file: {item['download_url']}")
                r = requests.get(item["download_url"])
                if r.status_code == 200:
                    with open(localPath, "wb") as f:
                        f.write(r.content)
                    print(f"[SUCCESS] Downloaded: {item['path']} -> {localPath}")
                else:
                    print(f"[ERROR] Failed to download file {item['path']}: HTTP {r.status_code}")
            except Exception as e:
                print(f"[EXCEPTION] While downloading file {item['path']}: {e}")
        elif item["type"] == "dir":
            print(f"[INFO] Entering directory: {item['path']}")
            downloadBranched(item["path"], localPath)
        else:
            print(f"[WARNING] Unknown item type: {item}")

def getFileJson(filePath):
    print(f"[INFO] Getting JSON file: {filePath}")
    try:
        text = getFileText(filePath)
        return json.loads(text)
    except Exception as e:
        print(f"[ERROR] getFileJson failed for {filePath}: {e}")
        raise

def getTextFromJson(json_obj):
    print(f"[INFO] Downloading content from JSON file object")
    if json_obj.get("type") != "file":
        raise Exception("Expected a file JSON object")
    try:
        r = requests.get(json_obj["download_url"])
        if r.status_code != 200:
            raise Exception(f"Failed to download file: {json_obj['path']} — Status {r.status_code}")
        return r.text
    except Exception as e:
        print(f"[ERROR] getTextFromJson failed: {e}")
        raise

def getDirChildrenJson(filePath):
    url = f"{API_BASE_URL}/{filePath}"
    print(f"[INFO] Fetching directory listing from: {url}")
    try:
        r = requests.get(url)
        if r.status_code != 200:
            raise Exception(f"Failed to list directory — Status code: {r.status_code}")
        return r.json()
    except Exception as e:
        print(f"[ERROR] getDirChildrenJson failed for {filePath}: {e}")
        raise

def findFirstChild(dirJson, name):
    print(f"[INFO] Looking for {name} in directory JSON...")
    for child in dirJson:
        if child["name"] == name:
            print(f"[FOUND] Found {name}")
            return child
    print(f"[WARNING] {name} not found in directory")
    return None

# Load config
try:
    configPath = os.path.join(os.path.dirname(os.path.abspath(__file__)), "config.json")
    print(f"[INFO] Loading config from: {configPath}")
    with open(configPath) as f:
        config = json.load(f)
except Exception as e:
    print(f"[FATAL] Failed to load config: {e}")
    sys.exit(1)

'''
CONFIG STRUCTURE
{
    "PackagesStorageDir": str
}
'''

if len(sys.argv) < 2:
    print("[USAGE] ulang install <package1> <package2> ...")
    sys.exit(1)

cmd = sys.argv[1]
args = sys.argv[2:]

print(f"[INFO] Command received: {cmd}")
match cmd:
    case "install":
        try:
            print(f"[INFO] Fetching registry metadata (data.json)")
            metaData = json.loads(getFileText("data.json"))
            packages = metaData["packages"]
        except Exception as e:
            print(f"[FATAL] Failed to load package registry: {e}")
            sys.exit(1)

        packageNames = packages
        for packageToInstall in args:
            print(f"[INFO] Installing package: {packageToInstall}")
            if packageToInstall not in packageNames:
                print(f"[ERROR] Package '{packageToInstall}' is not found in registry")
                continue
            try:
                pathToPackage = packageToInstall
                print(f"[INFO] Getting package directory listing for: {pathToPackage}")
                packageDir = getDirChildrenJson(pathToPackage)

                packageMetaData = findFirstChild(packageDir, "metadata.json")
                if not packageMetaData:
                    print(f"[ERROR] metadata.json not found in {packageToInstall}")
                    continue

                print(f"[INFO] Downloading metadata.json for {packageToInstall}")
                packageDICT = json.loads(getTextFromJson(packageMetaData))

                latestVer = packageDICT["latest"]
                latestVerPath = packageDICT["verstions"][latestVer]
                print(f"[INFO] Latest version is {latestVer}, path: {latestVerPath}")

                print(f"[INFO] Starting recursive download for version: {latestVer}")
                downloadBranched(latestVerPath, os.path.join(config["PackagesStorageDir"], packageDICT["Name"]))
            except Exception as e:
                print(f"[ERROR] Failed to install package '{packageToInstall}': {e}")
    case _:
        print(f"[ERROR] Unknown command: {cmd}")
