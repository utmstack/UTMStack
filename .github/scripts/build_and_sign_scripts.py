import os
import json
import subprocess
import requests
import shutil
import sys

def main():
    try:
        print("Script started.")

        # Set environment variables
        services = os.environ.get('SCRIPT_SERVICES', '').split(',')
        workspace = os.environ.get('GITHUB_WORKSPACE', '')
        replace_key = os.environ.get('AGENT_SECRET_PREFIX', '')
        sign_cert = os.environ.get('SIGN_CERT', '')
        sign_key = os.environ.get('SIGN_KEY', '')
        sign_container = os.environ.get('SIGN_CONTAINER', '')
        url = os.environ.get('CM_API', '')

        print(f"Services: {services}")
        print(f"Workspace: {workspace}")
        print(f"URL: {url}")

        # Parse JSON from CM_PUB_AUTH
        cm_pub_auth_json = os.environ.get('CM_PUB_AUTH', '{}')
        try:
            auth = json.loads(cm_pub_auth_json)
            print(f"Auth: {auth}")
        except json.JSONDecodeError as e:
            print(f"Failed to parse CM_PUB_AUTH JSON: {e}")
            raise

        # Load versions.json
        versions_json_path = os.path.join(workspace, 'versions.json')
        print(f"Loading versions.json from: {versions_json_path}")
        try:
            with open(versions_json_path, 'r') as f:
                versions_content = json.load(f)
            print(f"Versions: {versions_content}")
        except FileNotFoundError:
            print(f"versions.json not found at {versions_json_path}")
            raise
        except json.JSONDecodeError as e:
            print(f"Failed to parse versions.json: {e}")
            raise

        # Set up headers
        print("Setting up headers...")
        headers = {
            "publisher-key": auth.get('key', ''),
            "publisher-id": auth.get('id', '')
        }
        print(f"Headers set: {headers}")

        # Upsert Master Version
        master_changelog_path = os.path.join(workspace, 'CHANGELOG.md')
        print(f"Loading master changelog from: {master_changelog_path}")
        try:
            with open(master_changelog_path, 'r', encoding='utf-8') as f:
                master_changelog = f.read()
            print(f"Master Changelog loaded successfully.")
        except FileNotFoundError:
            print(f"CHANGELOG.md not found at {master_changelog_path}")
            raise

        print("Preparing master version data...")
        master_version_data = {
            "changelog": master_changelog,
            "version_name": versions_content.get('version', '')
        }

        print("Master version data prepared.")
        print("Attempting to upsert Master Version...")

        try:
            response = requests.post(
                f"{url}/master-version",
                headers=headers,
                json=master_version_data,
                timeout=300
            )
            response.raise_for_status()
            master_version_id = response.text.strip().strip('"')  # Removing any surrounding quotes
            print(f"Master Version ID: {master_version_id}")
        except requests.exceptions.RequestException as e:
            print(f"Error in Master Version POST request: {e}")
            if e.response is not None:
                print(f"Response Content: {e.response.text}")
            raise

        print("Master Version upserted successfully.")

        # Upsert Component Version, Scripts, and Files
        for service in services:
            print(f"Processing service: {service}")
            path = service.replace("-", "/")
            service_path = os.path.join(workspace, path)

            # Add specific pre-build steps here
            if service == "agent-service":
                config_path = os.path.join(service_path, 'config')
                print(f"Navigating to config directory: {config_path}")
                if not os.path.isdir(config_path):
                    print(f"Config directory not found: {config_path}")
                    raise FileNotFoundError(f"Config directory not found: {config_path}")
                os.chdir(config_path)
                const_go_path = os.path.join(config_path, 'const.go')
                print(f"Modifying const.go at {const_go_path}")
                try:
                    with open(const_go_path, 'r') as f:
                        const_go_content = f.read()
                    const_go_content = const_go_content.replace(
                        'const REPLACE_KEY string = ""',
                        f'const REPLACE_KEY string = "{replace_key}"'
                    )
                    with open(const_go_path, 'w') as f:
                        f.write(const_go_content)
                    print("const.go updated successfully.")
                except FileNotFoundError:
                    print(f"const.go not found at {const_go_path}")
                    raise
                except Exception as e:
                    print(f"Error modifying const.go: {e}")
                    raise

            os.chdir(service_path)
            print(f"Changed directory to: {service_path}")

            # Upsert Component Version
            changelog_path = os.path.join(service_path, 'CHANGELOG.md')
            readme_path = os.path.join(service_path, 'README.md')
            files_path = os.path.join(service_path, 'files.json')

            print(f"Loading component changelog from: {changelog_path}")
            try:
                with open(changelog_path, 'r', encoding='utf-8') as f:
                    changelog = f.read()
                print("Component changelog loaded successfully.")
            except FileNotFoundError:
                print(f"Component CHANGELOG.md not found at {changelog_path}")
                raise

            print(f"Loading README.md from: {readme_path}")
            try:
                with open(readme_path, 'r', encoding='utf-8') as f:
                    readme = f.read()
                print("README.md loaded successfully.")
            except FileNotFoundError:
                print(f"README.md not found at {readme_path}")
                raise

            component_version_data = {
                "changelog": changelog,
                "description": readme,
                "editions": ["Community", "Enterprise"],
                "master_version_id": master_version_id,
                "name": service,
                "version_name": versions_content.get(service, '')
            }

            print("Upserting Component Version...")
            try:
                response = requests.post(
                    f"{url}/component-version",
                    headers=headers,
                    json=component_version_data,
                    timeout=300
                )
                response.raise_for_status()
                component_version_id = response.text.strip().strip('"')  # Removing any surrounding quotes
                print(f"Component Version ID: {component_version_id}")
            except requests.exceptions.RequestException as e:
                print(f"Error in Component Version POST request: {e}")
                if e.response is not None:
                    print(f"Response Content: {e.response.text}")
                raise

            # Upsert Files and dependencies
            print(f"Loading files.json from: {files_path}")
            try:
                with open(files_path, 'r', encoding='utf-8') as f:
                    files_content = json.load(f)
                print("files.json loaded successfully.")
            except FileNotFoundError:
                print(f"files.json not found at {files_path}")
                raise
            except json.JSONDecodeError as e:
                print(f"Failed to parse files.json: {e}")
                raise

            for file in files_content:
                file_name = file.get('name', '')
                if not file_name:
                    print("File entry missing 'name'. Skipping...")
                    continue

                if file_name.endswith(('.zip', '.yaml', '.yml', '.json')):
                    file_path = os.path.join(service_path, 'dependencies', file_name)
                    print(f"Dependency file detected. File path: {file_path}")
                else:
                    # Build and sign the file if it's not a dependency
                    os.environ['GOOS'] = file.get('os', '')
                    os.environ['GOARCH'] = file.get('arch', '')
                    # Optionally disable Cgo for cross-compiling
                    os.environ['CGO_ENABLED'] = '0'
                    print(f"Building {file_name} for {os.environ['GOOS']} {os.environ['GOARCH']} with CGO_ENABLED={os.environ['CGO_ENABLED']}...")

                    # Fetch dependencies before building
                    print("Fetching Go module dependencies...")
                    mod_command = ['go', 'mod', 'tidy']
                    try:
                        mod_result = subprocess.run(mod_command, check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
                        print(f"Go modules fetched successfully. Output:\n{mod_result.stdout}")
                    except subprocess.CalledProcessError as e:
                        print(f"Error fetching Go modules: {e.stderr}")
                        raise

                    # Run 'go build' command
                    build_command = ['go', 'build', '-o', file_name, '-v', '.']
                    try:
                        result = subprocess.run(build_command, check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
                        print(f"Built {file_name} successfully. Output:\n{result.stdout}")
                    except subprocess.CalledProcessError as e:
                        print(f"Error building {file_name}: {e.stderr}")
                        raise

                    sign_key_str = "{{" + sign_key + "}}"

                    if os.environ['GOOS'].lower() == 'windows':
                        print(f"Signing {file_name}...")
                        # Run 'signtool' command
                        signtool_command = [
                            'signtool',
                            'sign',
                            '/fd', 'SHA256',
                            '/tr', 'http://timestamp.digicert.com',
                            '/td', 'SHA256',
                            '/f', sign_cert,
                            '/csp', 'eToken Base Cryptographic Provider',
                            '/k', f"[{{{{{sign_key}}}}}]={sign_container}",
                            file_name
                        ]
                        try:
                            result = subprocess.run(signtool_command, check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
                            print(f"Signed {file_name} successfully. Output:\n{result.stdout}")
                        except subprocess.CalledProcessError as e:
                            print(f"Error signing {file_name}: {e.stderr}")
                            raise

                    file_path = os.path.join(service_path, file_name)
                    print(f"Final file path: {file_path}")

                # Prepare file data
                file_data = {
                    "version_id": component_version_id,
                    "name": file_name,
                    "is_binary": file.get('is_binary', False),
                    "destination_path": file.get('destination_path', ''),
                    "os": file.get('os', ''),
                    "arch": file.get('arch', ''),
                    "replace_previous": file.get('replace_previous', False)
                }

                # Prepare multipart form data
                print(f"Uploading {file_name}...")
                try:
                    with open(file_path, 'rb') as f:
                        files = {
                            'data': (None, json.dumps(file_data), 'application/json'),
                            'file': (file_name, f, 'application/octet-stream')
                        }

                        response = requests.post(
                            f"{url}/upload-file",
                            headers=headers,
                            files=files,
                            timeout=300
                        )
                        response.raise_for_status()
                        response_content = response.text
                        print(f"Upload Response for {file_name}: {response_content}")
                except FileNotFoundError:
                    print(f"File not found for upload: {file_path}")
                    raise
                except requests.exceptions.RequestException as e:
                    print(f"Error uploading {file_name}: {e}")
                    if e.response is not None:
                        print(f"Response Content: {e.response.text}")
                    raise

        # Return to workspace directory
        print(f"Changing directory back to workspace: {workspace}")
        os.chdir(workspace)

    except Exception as e:
        print(f"Script failed: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()
