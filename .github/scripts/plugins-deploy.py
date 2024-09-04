import argparse
import json
import os
from google.cloud import storage

def main(environment):
	print("Deploying plugins to Google Cloud Storage")
	if environment.startswith("v10-"):
		environment = environment.replace("v10-", "")
		
	gcp_key = json.loads(os.environ.get("GCP_KEY"))
	storage_client = storage.Client.from_service_account_info(gcp_key)

	bucket_name = "utmstack-updates"
	bucket = storage_client.bucket(bucket_name)
	endpPlugins = environment + "/plugins/"
	endPipeline = environment + "/pipeline/"
	
	plugin_alerts_blob = bucket.blob(endpPlugins + "com.utmstack.alerts.plugin")
	plugin_events_blob = bucket.blob(endpPlugins + "com.utmstack.events.plugin")
	plugin_geolocation_blob = bucket.blob(endpPlugins + "com.utmstack.geolocation.plugin")
	plugin_stats_blob = bucket.blob(endpPlugins + "com.utmstack.stats.plugin")
	plugin_inputs_blob = bucket.blob(endpPlugins + "com.utmstack.inputs.plugin")
	plugin_gcp_blob = bucket.blob(endpPlugins + "com.utmstack.gcp.plugin")
	plugin_aws_blob = bucket.blob(endpPlugins + "com.utmstack.aws.plugin")
	plugin_azure_blob = bucket.blob(endpPlugins + "com.utmstack.azure.plugin")
	plugin_sophos_blob = bucket.blob(endpPlugins + "com.utmstack.sophos.plugin")
	plugin_o365_blob = bucket.blob(endpPlugins + "com.utmstack.o365.plugin")
	plugin_config_blob = bucket.blob(endpPlugins + "com.utmstack.config.plugin")
	plugin_bitdefender_blob = bucket.blob(endpPlugins + "com.utmstack.bitdefender.plugin")

	pipeline_analysis_blob = bucket.blob(endPipeline + "system_plugins_analysis.yaml")
	pipeline_correlation_blob = bucket.blob(endPipeline + "system_plugins_correlation.yaml")
	pipeline_input_blob = bucket.blob(endPipeline + "system_plugins_input.yaml")
	pipeline_notification_blob = bucket.blob(endPipeline + "system_plugins_notification.yaml")
	pipeline_parsing_blob = bucket.blob(endPipeline + "system_plugins_parsing.yaml")
	
	plugin_alerts_blob.upload_from_filename(os.path.join(os.environ["GITHUB_WORKSPACE"], "plugins", "alerts", "com.utmstack.alerts.plugin"))
	plugin_events_blob.upload_from_filename(os.path.join(os.environ["GITHUB_WORKSPACE"], "plugins", "events", "com.utmstack.events.plugin"))
	plugin_geolocation_blob.upload_from_filename(os.path.join(os.environ["GITHUB_WORKSPACE"], "plugins", "geolocation", "com.utmstack.geolocation.plugin"))
	plugin_stats_blob.upload_from_filename(os.path.join(os.environ["GITHUB_WORKSPACE"], "plugins", "stats", "com.utmstack.stats.plugin"))
	plugin_inputs_blob.upload_from_filename(os.path.join(os.environ["GITHUB_WORKSPACE"], "plugins", "inputs", "com.utmstack.inputs.plugin"))
	plugin_gcp_blob.upload_from_filename(os.path.join(os.environ["GITHUB_WORKSPACE"], "plugins", "gcp", "com.utmstack.gcp.plugin"))	
	plugin_aws_blob.upload_from_filename(os.path.join(os.environ["GITHUB_WORKSPACE"], "plugins", "aws", "com.utmstack.aws.plugin"))
	plugin_azure_blob.upload_from_filename(os.path.join(os.environ["GITHUB_WORKSPACE"], "plugins", "azure", "com.utmstack.azure.plugin"))
	plugin_sophos_blob.upload_from_filename(os.path.join(os.environ["GITHUB_WORKSPACE"], "plugins", "sophos", "com.utmstack.sophos.plugin")) 
	plugin_o365_blob.upload_from_filename(os.path.join(os.environ["GITHUB_WORKSPACE"], "plugins", "o365", "com.utmstack.o365.plugin"))	
	plugin_config_blob.upload_from_filename(os.path.join(os.environ["GITHUB_WORKSPACE"], "plugins", "config", "com.utmstack.config.plugin"))	
	plugin_bitdefender_blob.upload_from_filename(os.path.join(os.environ["GITHUB_WORKSPACE"], "plugins", "bitdefender", "com.utmstack.config.plugin"))	

	pipeline_analysis_blob.upload_from_filename(os.path.join(os.environ["GITHUB_WORKSPACE"], "etc", "threatwinds", "pipeline", "system_plugins_analysis.yaml"))
	pipeline_correlation_blob.upload_from_filename(os.path.join(os.environ["GITHUB_WORKSPACE"], "etc", "threatwinds", "pipeline", "system_plugins_correlation.yaml"))
	pipeline_input_blob.upload_from_filename(os.path.join(os.environ["GITHUB_WORKSPACE"], "etc", "threatwinds", "pipeline", "system_plugins_input.yaml"))
	pipeline_notification_blob.upload_from_filename(os.path.join(os.environ["GITHUB_WORKSPACE"], "etc", "threatwinds", "pipeline", "system_plugins_notification.yaml"))
		
if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Update UTMStack Dependencies in Google Cloud Storage")
    parser.add_argument("environment", type=str, help="Environment(dev, qa, rc, release)")
    
    args = parser.parse_args()
    main(args.environment)