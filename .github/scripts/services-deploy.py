import argparse
import json
import os
import requests
from google.cloud import storage
import yaml

def main(environment, depend):
	if environment.startswith("v10-"):
		environment = environment.replace("v10-", "")
		
	gcp_key = json.loads(os.environ.get("GCP_KEY"))
	storage_client = storage.Client.from_service_account_info(gcp_key)

	bucket_name = "utmstack-updates"
	bucket = storage_client.bucket(bucket_name)

	# Read utmstack version from version.yml
	with open(os.path.join(os.environ["GITHUB_WORKSPACE"], "version.yml"), "r") as f:
		local_master_version = yaml.safe_load(f)['version']

	# Read services version from version.json
	with open(os.path.join(os.environ["GITHUB_WORKSPACE"], depend, "version.json"), "r") as f:
		local_versions = json.load(f)

	# Download version.json from URL
	endp = environment + "/" + depend.replace("-", "_")
	response = requests.get("https://storage.googleapis.com/" + bucket_name + "/" + endp + "/version.json")
	remote_versions = response.json()['versions']

	# Find the object with matching master_version
	version_found = False
	for obj in remote_versions:
		if obj['master_version'] == local_master_version:
			if "installer_version" in local_versions:
				obj['installer_version'] = local_versions["installer_version"]
			if "service_version" in local_versions:
				obj['service_version'] = local_versions["service_version"]
			if "dependencies_version" in local_versions:
				obj['dependencies_version'] = local_versions["dependencies_version"]
			version_found = True
			break

	# If no matching object found, create a new one
	if version_found == False:
		version_obj = {
			'master_version': local_master_version,
		}
		if "installer_version" in local_versions:
			version_obj['installer_version'] = local_versions["installer_version"]
		if "service_version" in local_versions:
			version_obj['service_version'] = local_versions["service_version"]
		if "dependencies_version" in local_versions:
			version_obj['dependencies_version'] = local_versions["dependencies_version"]
		remote_versions.append(version_obj)
		
	# Update version.json
	version_blob = bucket.blob(endp + "/version.json")
	version_blob.upload_from_string(json.dumps({'versions': remote_versions}, indent=4))
	
	# Create blobs and upload services
	if "service_version" in local_versions:
		if depend == "agent":
			service_windows_blob = bucket.blob(endp + "/utmstack_agent_service_v" + local_versions["service_version"].replace(".","_") + "_windows.exe")
			service_linux_blob = bucket.blob(endp + "/utmstack_agent_service_v" + local_versions["service_version"].replace(".","_") + "_linux")
			service_windows_blob.upload_from_filename(os.path.join(os.environ["GITHUB_WORKSPACE"], "agent", "agent", "utmstack_agent_service.exe"))
			service_linux_blob.upload_from_filename(os.path.join(os.environ["GITHUB_WORKSPACE"], "agent", "agent", "utmstack_agent_service"))
	if "installer_version" in local_versions:
		if depend == "agent":
			installer_windows_blob = bucket.blob(endp + "/utmstack_agent_installer_v" + local_versions["installer_version"].replace(".","_") + "_windows.exe")
			installer_linux_blob = bucket.blob(endp + "/utmstack_agent_installer_v" + local_versions["installer_version"].replace(".","_") + "_linux")
			installer_windows_blob.upload_from_filename(os.path.join(os.environ["GITHUB_WORKSPACE"], "agent", "installer", "utmstack_agent_installer.exe"))
			installer_linux_blob.upload_from_filename(os.path.join(os.environ["GITHUB_WORKSPACE"], "agent", "installer", "utmstack_agent_installer"))
		if depend == "collector-installer":
			installer_windows_blob = bucket.blob(endp + "/utmstack_collector_installer_v" + local_versions["installer_version"].replace(".","_") + "_windows.exe")
			installer_linux_blob = bucket.blob(endp + "/utmstack_collector_installer_v" + local_versions["installer_version"].replace(".","_") + "_linux")
			installer_windows_blob.upload_from_filename(os.path.join(os.environ["GITHUB_WORKSPACE"], depend, "utmstack_collector_installer.exe"))
			installer_linux_blob.upload_from_filename(os.path.join(os.environ["GITHUB_WORKSPACE"], depend, "utmstack_collector_installer"))

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Update UTMStack Dependencies in Google Cloud Storage")
    parser.add_argument("environment", type=str, help="Environment(dev, qa, rc, release)")
    parser.add_argument("depend", type=str, help="Dependencies to update(agent, collector-installer)")
    
    args = parser.parse_args()
    main(args.environment, args.depend)