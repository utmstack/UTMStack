import argparse
import json
import os
import requests
from google.cloud import storage
import yaml

def main(environment):
	gcp_key = json.loads(os.environ.get("GCP_KEY"))
	storage_client = storage.Client.from_service_account_info(gcp_key)

	bucket_name = "utmstack-updates"
	bucket = storage_client.bucket(bucket_name)

	# Read utmstack version from version.yml
	with open(os.path.join(os.environ["GITHUB_WORKSPACE"], "version.yml"), "r") as f:
		local_master_version = yaml.safe_load(f)['version']

	# Read agent services version from versions.json
	with open(os.path.join(os.environ["GITHUB_WORKSPACE"], "agent", "versions.json"), "r") as f:
		local_agent_versions = json.load(f)

	# Download versions.json from URL
	endp = "agent_updates/{}".format(environment)
	response = requests.get(endp + "/versions.json")
	remote_versions = response.json()['versions']

	# Find the object with matching master_version
	version_found = False
	for obj in remote_versions:
		if obj['master_version'] == local_master_version:
			obj['agent_version'] = local_agent_versions["agent_version"]
			obj['updater_version'] = local_agent_versions["updater_version"]
			obj['redline_version'] = local_agent_versions["redline_version"]
			version_found = True
			break

	# If no matching object found, create a new one
	if version_found == False:
		version_obj = {
			'master_version': local_master_version,
			'agent_version': local_agent_versions["agent_version"],
			'updater_version': local_agent_versions["updater_version"],
			'redline_version': local_agent_versions["redline_version"],
		}
		remote_versions.append(version_obj)
		
	# Update version.json
	version_blob = bucket.blob(endp + "/versions.json")
	version_blob.upload_from_string(json.dumps({'versions': remote_versions}, indent=4))
	
	# Create agent blobs
	agent_windows_blob = bucket.blob(endp + "/agent_service/v" + local_agent_versions["agent_version"] + "/utmstack_agent_service.exe")
	agent_linux_blob = bucket.blob(endp + "/agent_service/v" + local_agent_versions["agent_version"] + "/utmstack_agent_service")
	updater_windows_blob = bucket.blob(endp + "/updater_service/v" + local_agent_versions["updater_version"] + "/utmstack_updater_service.exe")
	updater_linux_blob = bucket.blob(endp + "/updater_service/v" + local_agent_versions["updater_version"] + "/utmstack_updater_service")
	redline_windows_blob = bucket.blob(endp + "/redline_service/v" + local_agent_versions["redline_version"] + "/utmstack_redline_service.exe")
	redline_linux_blob = bucket.blob(endp + "/redline_service/v" + local_agent_versions["redline_version"] + "/utmstack_redline_service")
	installer_windows_blob = bucket.blob(endp + "/installer/v" + local_master_version + "/utmstack_agent_installer.exe")
	installer_linux_blob = bucket.blob(endp + "/installer/v" + local_master_version + "/utmstack_agent_installer")

	# Upload agent services
	agent_windows_blob.upload_from_filename(os.path.join(os.environ["GITHUB_WORKSPACE"], "agent", "agent_service", "utmstack_agent_service.exe"))
	agent_linux_blob.upload_from_filename(os.path.join(os.environ["GITHUB_WORKSPACE"], "agent", "agent_service", "utmstack_agent_service"))
	updater_windows_blob.upload_from_filename(os.path.join(os.environ["GITHUB_WORKSPACE"], "agent", "updater_service", "utmstack_updater_service.exe"))
	updater_linux_blob.upload_from_filename(os.path.join(os.environ["GITHUB_WORKSPACE"], "agent", "updater_service", "utmstack_updater_service"))
	redline_windows_blob.upload_from_filename(os.path.join(os.environ["GITHUB_WORKSPACE"], "agent", "redline_service", "utmstack_redline_service.exe"))
	redline_linux_blob.upload_from_filename(os.path.join(os.environ["GITHUB_WORKSPACE"], "agent", "redline_service", "utmstack_redline_service"))
	installer_windows_blob.upload_from_filename(os.path.join(os.environ["GITHUB_WORKSPACE"], "agent", "installer", "utmstack_agent_installer.exe"))
	installer_linux_blob.upload_from_filename(os.path.join(os.environ["GITHUB_WORKSPACE"], "agent", "installer", "utmstack_agent_installer"))

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Update UTMStack agents in Google Cloud Storage")
    parser.add_argument("environment", type=str, help="Environment(dev, qa, rc, release)")
    
    args = parser.parse_args()
    main(args.environment)