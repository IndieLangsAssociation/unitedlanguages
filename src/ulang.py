import json
import sys, os, requests

RAW_BASE_URL = 'https://raw.githubusercontent.com/IndieLangsAssociation/packages/main'
API_BASE_URL = 'https://api.github.com/repos/IndieLangsAssociation/packages/contents'

def getFileText(filePath):
    url = f"{RAW_BASE_URL}/{filePath}"
    r = requests.get(url)
    if r.status_code != 200:
        raise Exception(f"Failed to fetch text from {url} — {r.status_code}")
    return r.text

def getFileJson(filePath):
    text = getFileText(filePath)
    return json.loads(text)

def getTextFromJson(json_obj):
    # Assumes the object is a GitHub contents API file item
    if json_obj.get("type") != "file":
        raise Exception("Expected a file JSON object")
    r = requests.get(json_obj["download_url"])
    if r.status_code != 200:
        raise Exception(f"Failed to download file: {json_obj['path']}")
    return r.text

def getDirChildrenJson(filePath):
    url = f"{API_BASE_URL}/{filePath}"
    r = requests.get(url)
    if r.status_code != 200:
        raise Exception(f"Failed to list directory: {url} — {r.status_code}")
    return r.json()

# Load config
configPath = os.path.join(os.path.dirname(os.path.abspath(__file__)), "config.json")
config = json.loads(open(configPath).read())

'''
CONFIG STRUCTURE
{
    "PackagesStorageDir": str
}
'''

cmd = sys.argv[1]
args = sys.argv[2:]

match cmd:
    case "install":
        metaData = json.loads(getFileText("data.json"))
        packages = metaData["packages"]
        packageNames = packages
        for packageToInstall in args:
            if not packageToInstall in packageNames:
                print(f"ERROR: package {packageToInstall} in not found in registory")
                exit()
            pathToPackage = packageToInstall
            packageDir = getDirChildrenJson(pathToPackage)
            for _ in packageDir: print(_)