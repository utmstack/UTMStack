import argparse
import json
import os
from google.cloud import storage

def main(update_version, file_type, binary):
	gcp_key = json.loads(os.environ.get("GCP_KEY"))
	storage_client = storage.Client.from_service_account_info(gcp_key)

	bucket_name = "utmstack-updates"
	bucket = storage_client.bucket(bucket_name)
    
	if file_type == "installer":
		endp = "agent_updates/release/installer/"
	else :
		endp = "agent_updates/release/" + file_type + "_service/"

    # Upload files
	windows_blob = bucket.blob(endp + update_version + "/" + binary +  ".exe")
	windows_blob.upload_from_filename(os.path.join(os.environ["GITHUB_WORKSPACE"], "agent", file_type, (binary+".exe")))

	linux_blob = bucket.blob(endp + update_version + "/" + binary)
	linux_blob.upload_from_filename(os.path.join(os.environ["GITHUB_WORKSPACE"],"agent", file_type, binary))

    # Update version.json
	version_blob = bucket.blob("agent_updates/release/versions.json")
	version_blob.upload_from_filename(os.path.join(os.environ["GITHUB_WORKSPACE"],"agent", "versions.json"))

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Update UTMStack agents in Google Cloud Storage")
    parser.add_argument("update_version", type=str, help="Update version string (v1.0.1)")
    parser.add_argument("file_type", type=str, help="File type string(agent, installer, redline, updater)")
    parser.add_argument("binary", type=str, help="Binary string(utmstack_agent_installer)")
    
    args = parser.parse_args()
    main(args.update_version, args.file_type, args.binary)