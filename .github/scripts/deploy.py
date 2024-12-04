import os
import json
import subprocess
import requests
import sys

def determine_environment(event_name, ref, review_state, base_ref, head_ref):
    """
    Determines the environment based on the GitHub event.
    """
    if event_name == "push" and ref.startswith("refs/heads/feature/"):
        return "dev"
    elif event_name == "pull_request_review" and review_state == "approved" and base_ref == "main" and head_ref.startswith("feature/"):
        return "qa"
    elif event_name == "push" and ref == "refs/heads/main":
        return "rc"
    elif event_name == "release":
        return "prod"
    else:
        return None

def set_output(name, value):
    """
    Writes output to the GITHUB_OUTPUT file so it can be used in later steps.
    """
    output_path = os.environ.get('GITHUB_OUTPUT')
    if output_path:
        with open(output_path, 'a') as fh:
            print(f'{name}={value}', file=fh)
    else:
        print(f"GITHUB_OUTPUT is not defined. Unable to set output {name}.")

def load_file(file_path, file_format, optional=False):
    """
    Loads a file and returns it in the specified format.
    """
    if not os.path.exists(file_path) or not os.path.isfile(file_path):
        if optional:
            print(f"{file_path} not found. Continuing as it is optional...")
            return {} if file_format == 'json' else ""
        else:
            print(f"{file_path} not found. Exiting...")
            sys.exit(1)
    with open(file_path, 'r', encoding='utf-8') as f:
        if file_format == 'json':
            try:
                content = json.load(f)
            except json.JSONDecodeError as e:
                print(f"Error decoding JSON: {e}")
                sys.exit(1)
        elif file_format == 'string':
            content = f.read()
        else:
            print(f"Unsupported file format: {file_format}. Exiting...")
            sys.exit(1)
    return content

def get_cm_credentials(env_tag):
    """
    Gets the CM credentials based on the environment.
    """
    cm_api_env = f'CM_API_{env_tag.upper()}'
    cm_pub_auth_env = f'CM_PUB_AUTH_{env_tag.upper()}'
    cm_api = os.getenv(cm_api_env)
    cm_pub_auth = os.getenv(cm_pub_auth_env)
    if cm_api is None or cm_pub_auth is None:
        print(f"CM credentials not found for {env_tag}. Exiting...")
        return None, None
    try:
        cm_pub_auth_json = json.loads(cm_pub_auth)
    except json.JSONDecodeError as e:
        print(f"Error decoding CM_PUB_AUTH: {e}")
        return None, None
    headers = {
        "publisher-key": cm_pub_auth_json.get("key"),
        "publisher-id": cm_pub_auth_json.get("id")
    }
    return cm_api, headers
    
def fetch_api_versions(url, headers):
    """
    Gets component versions from the CM API.
    """
    try:
        response = requests.get(url, headers=headers, verify=False)
        response.raise_for_status()
        return response.json()
    except requests.RequestException as e:
        print(f"Failed to fetch API versions: {e}")
        return {}

def compare_versions(cm_server, headers, local_versions, master_version):
    """
    Compare local versions with API versions to identify updated services.
    """
    file_based_services = ["agent-service", "agent-installer", "pipelines", "plugins-alerts", "plugins-aws", "plugins-azure", "plugins-bitdefender", "plugins-config", "plugins-events", "plugins-gcp", "plugins-geolocation", "plugins-inputs", "plugins-o365", "plugins-sophos", "plugins-stats"]
    image_based_services = ["agent-manager", "backend", "frontend", "user-auditor", "web-pdf"]

    updated_file_based_services=[]
    updated_image_based_services=[]

    api_url = f"{cm_server}/component-versions?master-version={master_version}"
    api_versions=fetch_api_versions(api_url, headers)

    for service, service_version in local_versions.items():
        if service == "master-version":
            continue
        api_version = api_versions.get(service)
        if service_version != api_version:
            if service in file_based_services:
                print(f"Update file based service detected: {service}")
                updated_file_based_services.append(service)
            elif service in image_based_services:
                print(f"Update image service detected: {service}")
                updated_image_based_services.append(service)
            else:
                print(f"Unknown service: {service}. Skipping...")
    return updated_file_based_services, updated_image_based_services

def prebuild_agent_service(service, service_path, replace_key):
    """
    Perform pre-build steps for 'agent-service'.
    """
    if service == "agent-service":
        const_go_path = os.path.join(service_path, 'config', 'const.go')
        if os.path.exists(const_go_path):
            print(f"Modifying {const_go_path}")
            with open(const_go_path, 'r', encoding='utf-8') as f:
                content = f.read()
            content = content.replace(
                'const REPLACE_KEY string = ""',
                f'const REPLACE_KEY string = "{replace_key}"'
            )
            with open(const_go_path, 'w', encoding='utf-8') as f:
                f.write(content)
            print("const.go updated successfully.")
        else:
            print(f"{const_go_path} not found. Skipping...")

def upsert_master_version(url, headers, master_version, workspace):
    """
    Push or update the master version in the CM.
    """
    print("Loading master changelog...")
    master_changelog = load_file(os.path.join(workspace, 'CHANGELOG.md'), 'string', True)
    master_version_data = {
        "changelog": master_changelog,
        "version_name": master_version
    }
    try:
        response = requests.post(
            f"{url}/master-version",
            headers=headers,
            json=master_version_data,
            verify=False,
            timeout=300
        )
        response.raise_for_status()
        return response.text.strip().strip('"')
    except requests.RequestException as e:
        print(f"Failed to upsert Master Version: {e}")
        sys.exit(1)

def upload_component_version(url, headers, service_path, master_version_id, service, service_version):
    """
    Pushes or updates the version of a component in the CM.
    """
    changelog = load_file(os.path.join(service_path, 'CHANGELOG.md'), 'string', True)
    readme = load_file(os.path.join(service_path, 'README.md'), 'string', True)
    component_version_data = {
        "changelog": changelog,
        "description": readme,
        "editions": ["Community", "Enterprise"],
        "master_version_id": master_version_id,
        "name": service,
        "version_name": service_version
    }
    try:
        response = requests.post(
            f"{url}/component-version",
            headers=headers,
            json=component_version_data,
            verify=False,
            timeout=300
        )
        response.raise_for_status()
        return response.text.strip().strip('"')
    except requests.RequestException as e:
        print(f"Failed to upsert Component Version: {e}")
        sys.exit(1)

def upload_files(url, headers, service_path, component_version_id, sign_cert, sign_key, sign_container):
    """
    Build, sign and upload files to CM.
    """
    files_content = load_file(os.path.join(service_path, 'files.json'), 'json', True)
    if not files_content:
        print("files.json not found. Skipping file upload...")
        return
    
    for file_info in files_content:
        file_name = file_info.get('name', '')
        if not file_name:
            print("File entry missing 'name'. Skipping...")
            continue

        filepath = os.path.join(service_path, file_name)

        if file_info.get('is_binary', False):
            # Build and sign the file if it's not a dependency
            os.environ['GOOS'] = file_info.get('os', '')
            os.environ['GOARCH'] = file_info.get('arch', '')
            os.environ['CGO_ENABLED'] = '0'
            print(f"Building {file_name} for {os.environ['GOOS']} {os.environ['GOARCH']} with CGO_ENABLED={os.environ['CGO_ENABLED']}...")

            # Fetch dependencies before building
            try:
                subprocess.run(['go', 'mod', 'tidy'], check=True, cwd=service_path)
            except subprocess.CalledProcessError as e:
                print(f"Failed to fetch dependencies: {e}")

            # Run 'go build' command
            try:
                subprocess.run(['go', 'build', '-o', file_name, '-v', '.'], check=True, cwd=service_path)
                print(f"Built {file_name} successfully.")
            except subprocess.CalledProcessError as e:
                print(f"Failed to build {file_name}: {e}")
                continue

            if os.environ['GOOS'].lower() == 'windows':
                print(f"Signing {file_name}...")
                sign_key_str = "{{" + sign_key + "}}"
                signtool_command = [
                    'signtool', 'sign', '/fd', 'SHA256', '/tr', 'http://timestamp.digicert.com',
                    '/td', 'SHA256', '/f', sign_cert, '/csp', 'eToken Base Cryptographic Provider',
                    '/k', f"[{{{{{sign_key}}}}}]={sign_container}", file_name
                ]
                try:
                    subprocess.run(signtool_command, check=True, cwd=service_path)
                    print(f"Signed {file_name} successfully.")
                except subprocess.CalledProcessError as e:
                    print(f"Failed to sign {file_name}: {e}")
                    continue

        # Prepare file data
        file_data = {
            "version_id": component_version_id,
            "name": file_name,
            "is_binary": file_info.get('is_binary', False),
            "destination_path": file_info.get('destination_path', ''),
            "os": file_info.get('os', ''),
            "arch": file_info.get('arch', ''),
            "replace_previous": file_info.get('replace_previous', False)
        }

        # Prepare multipart form data
        print(f"Uploading {file_name}...")
        with open(filepath, 'rb') as f:
            files = {
                'data': (None, json.dumps(file_data), 'application/json'),
                'file': (file_name, f, 'application/octet-stream')
            }

            try:
                response = requests.post(
                    f"{url}/upload-file",
                    headers=headers,
                    files=files,
                    verify=False,
                    timeout=300
                )
                response.raise_for_status()
                print(f"{file_name} uploaded successfully.")
            except requests.RequestException as e:
                print(f"Failed to upload {file_name}: {e}")

def upload_scripts(url, headers, service_path, component_version_id, version, env_tag):
    """
    Upload scripts associated with a component to the CM.
    """
    scripts_content = load_file(os.path.join(service_path, 'scripts.json'), 'json', True)
    if not scripts_content:
        print("scripts.json not found. Skipping script upload...")
        return
    
    replacement = f"v{version}-{env_tag}"
    for script in scripts_content:
        if 'imagetagreplace' in script:
            script = script.replace('imagetagreplace', replacement)

        script_data = {
            "version_id": component_version_id,
            "script": script,
        }
        print(f"Uploading {script}...")
        try:
            response = requests.post(
                f"{url}/script",
                headers=headers,
                json=script_data,
                verify=False,
                timeout=300
            )
            response.raise_for_status()
            print(f"{script} uploaded successfully.")
        except requests.RequestException as e:
            print(f"Failed to upload {script}: {e}")

def upload_file_based_services(url, headers, services, versions_content, master_version_id, workspace, replace_key, sign_cert, sign_key, sign_container, env_tag):
    """
    Process and upload file-based services.
    """
    for service in services:
        print(f"Processing file based service: {service}")
        service_path = os.path.join(workspace, service.replace("-", "/"))

        # Add specific pre-build steps here
        prebuild_agent_service(service, service_path, replace_key)

        os.chdir(service_path)

        component_version_id = upload_component_version(
            url, headers, service_path, master_version_id, service, versions_content.get(service, ''))
        print(f"Component Version ID: {component_version_id}")

        print("Uploading files...")
        upload_files(url, headers, service_path, component_version_id, sign_cert, sign_key, sign_container)
        print(f"Files uploaded successfully for {service}")

        print("Uploading scripts...") 
        upload_scripts(url, headers, service_path, component_version_id, versions_content.get(service, ''),env_tag)
        print(f"Scripts uploaded successfully for {service}")

def upload_image_based_services(url, headers, services, versions_content, master_version_id, workspace,env_tag):
    """
    Process and upload image-based services.
    """
    for service in services:
        print(f"Processing image based service: {service}")
        service_path = os.path.join(workspace, service)

        os.chdir(service_path)

        component_version_id = upload_component_version(url, headers, service_path, master_version_id, service, versions_content.get(service, ''))
        print(f"Component Version ID: {component_version_id}")

        print("Uploading scripts...") 
        upload_scripts(url, headers, service_path, component_version_id, versions_content.get(service, ''),env_tag)
        print(f"Scripts uploaded successfully for {service}")   

def main():
    event_name = os.getenv('GITHUB_EVENT_NAME')
    ref = os.getenv('GITHUB_REF')
    review_state = os.getenv('GITHUB_REVIEW_STATE')
    base_ref = os.getenv('GITHUB_BASE_REF')
    head_ref = os.getenv('GITHUB_HEAD_REF')
    workspace = os.getenv('GITHUB_WORKSPACE')
    replace_key = os.getenv('AGENT_SECRET_PREFIX')
    sign_cert = os.getenv('SIGN_CERT')
    sign_key = os.getenv('SIGN_KEY')
    sign_container = os.getenv('SIGN_CONTAINER')

    print("Determining environment...")
    env_tag = determine_environment(event_name, ref, review_state, base_ref, head_ref)
    if env_tag is None:
        print("Environment not determined. Exiting...")
        sys.exit(1)
    set_output("env_tag", env_tag)
    print(f"Environment: {env_tag}")

    print("Getting versions from versions.json...")
    versions = load_file(os.path.join(workspace, 'versions.json'), 'json')
    master_version = versions.get('master-version')
    print(f"Versions: {versions}")
    print(f"Master Version: {master_version}")

    print("Fetching CM credentials...")
    cm_server, headers = get_cm_credentials(env_tag)
    if cm_server is None or headers is None:
        print("CM credentials not found")
        sys.exit(1)

    print("Comparing api versions and file versions...")
    updated_file_based_services, updated_image_based_services = compare_versions(cm_server, headers, versions, master_version)
    set_output("image_based_services",updated_image_based_services)

    print("Attempting to upsert Master Version...")
    master_version_id = upsert_master_version(cm_server, headers, master_version, workspace)
    print(f"Master Version ID: {master_version_id}")

    print("Uploading file based services...")
    upload_file_based_services(
        cm_server, headers, updated_file_based_services, versions, master_version_id,
        workspace, replace_key, sign_cert, sign_key, sign_container, env_tag
    )
    print("File based services uploaded successfully")

    print("Uploading image based services...")
    upload_image_based_services(cm_server, headers, updated_image_based_services, versions, master_version_id, workspace,env_tag)
    print("Image based services uploaded successfully")

if __name__ == "__main__":
    main()
