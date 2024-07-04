import argparse
import json
import os
import requests
from google.cloud import storage
import yaml

def main(environment):
	if environment.startswith("v10-"):
		environment = environment.replace("v10-", "")
		
	gcp_key = json.loads(os.environ.get("GCP_KEY"))
	storage_client = storage.Client.from_service_account_info(gcp_key)

	bucket_name = "utmstack-updates"
	bucket = storage_client.bucket(bucket_name)
	endp = environment + "/plugins/"
	
	alerts_blob = bucket.blob(endp + "com.utmstack.alerts.plugin")
	events_blob = bucket.blob(endp + "com.utmstack.events.plugin")
	geolocation_blob = bucket.blob(endp + "com.utmstack.geolocation.plugin")
	inputs_blob = bucket.blob(endp + "com.utmstack.inputs.plugin")
	
	alerts_blob.upload_from_filename(os.path.join(os.environ["GITHUB_WORKSPACE"], "plugins", "alerts", "com.utmstack.alerts.plugin"))
	events_blob.upload_from_filename(os.path.join(os.environ["GITHUB_WORKSPACE"], "plugins", "events", "com.utmstack.events.plugin"))
	geolocation_blob.upload_from_filename(os.path.join(os.environ["GITHUB_WORKSPACE"], "plugins", "geolocation", "com.utmstack.geolocation.plugin"))
	inputs_blob.upload_from_filename(os.path.join(os.environ["GITHUB_WORKSPACE"], "plugins", "inputs", "com.utmstack.inputs.plugin"))
		
if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Update UTMStack Dependencies in Google Cloud Storage")
    parser.add_argument("environment", type=str, help="Environment(dev, qa, rc, release)")
    
    args = parser.parse_args()
    main(args.environment)