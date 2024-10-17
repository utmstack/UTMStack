import argparse
import json
import os
import shutil
import requests
from datetime import datetime
import yaml

def parse_readme(depend):
	with open(os.path.join(os.environ["GITHUB_WORKSPACE"], depend, "README.md"), "r", encoding='utf-8') as f:
		lines = f.readlines()
		if lines[0].startswith("# "):
			name = lines[0].replace("# ", "").strip().replace(" ", "")
		description_lines = []
		for line in lines[1:]:
			if line.startswith("# "):
				break
			description_lines.append(line.strip())
		description = " ".join(description_lines)
		return name, description
	
def download_file(url, local_filename):
    with requests.get(url, stream=True) as r:
        r.raise_for_status()
        with open(local_filename, 'wb') as f:
            shutil.copyfileobj(r.raw, f)

def main(environment, depend):
	publisher_key_json = json.loads(os.environ.get("PUBLISHER_KEY"))
	cm_server = os.environ.get("CM_SERVER")
	
	with open(os.path.join(os.environ["GITHUB_WORKSPACE"], depend, "version.yml"), "r") as f:
		local_service_version = yaml.safe_load(f)['version']

	with open(os.path.join(os.environ["GITHUB_WORKSPACE"], depend, "CHANGELOG.md"), "r") as f:
		changelog = f.read()

	with open(os.path.join(os.environ["GITHUB_WORKSPACE"], depend, "files.json"), "r") as f:
		files = json.load(f)

	name, description = parse_readme(depend)

	data = {
		"version": local_service_version,
		"build_number": int(datetime.now().strftime("%Y%m%d%H%M")),
		"changelog": changelog,
		"component": {
			"name": name,
			"description": description,
			"types": ["Community", "Enterprise"],
		},
		"files": files
	}

	data_json = json.dumps(data)

	filesMap = {}
	for file in files:
		if file["name"].endswith(".zip"):
			download_file("https://storage.googleapis.com/utmstack-updates/" + environment + "/" + depend +"/" + file["name"], os.path.join(os.environ["GITHUB_WORKSPACE"], "build", file["name"]))
		filesMap[file["name"]] = open(os.path.join(os.environ["GITHUB_WORKSPACE"], "build", file["name"]))

	form_data = {"data": data_json}

	response = requests.post(cm_server, data=form_data, files=filesMap, headers={"publisher-key": publisher_key_json["key"], "publisher-id": publisher_key_json["id"]})
	
	print(response.text)

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Update UTMStack Dependencies in Customer Manager")
    parser.add_argument("environment", type=str, help="Environment(dev, qa, rc, prod)")
    parser.add_argument("depend", type=str, help="Dependency to update(agent, collector-installer)")

    args = parser.parse_args()
    main(args.environment, args.depend)