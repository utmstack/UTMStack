import json
import os
import time
from string import Template

import jwt
import requests

from cloud_integrations.integration import Integration
from util.aes_cipher import AESCipher
from util.module_enum import GOOGLE_PUBSUB
from util.misc import get_module_group

LOCATION = "/etc/utmstack/"

SCOPE = "https://www.googleapis.com/auth/cloud-platform"

# Set how long this token will be valid in seconds
EXPIRE = 3600  # Expires in 1 hour


def get_google_key_path(json_key_file, project_id) -> str:
    """Return json key location location"""
    try:
        file = LOCATION + project_id + ".json"
        directory = os.path.dirname(LOCATION)
        if not os.path.exists(directory):
            os.makedirs(directory)
        with open(os.open(file, os.O_CREAT | os.O_WRONLY, 0o777), 'w') as outfile:
            json.dump(json_key_file, outfile)
        outfile.close()
        return file
    except Exception as exception:
        print('Error saving jsonKey :' + str(exception))


def load_json_credentials(filename):
    """ Load the Google Service Account Credentials from Json file """
    with open(filename, 'r') as f:
        data = f.read()
    return json.loads(data)


def load_private_key(json_cred):
    """ Return the private key from the json credentials """
    return json_cred['private_key']


def create_signed_jwt(pkey, pkey_id, email, scope):
    """
    Create a Signed JWT from a service account Json credentials file
    This Signed JWT will later be exchanged for an Access Token
    """

    # Google Endpoint for creating OAuth 2.0 Access Tokens from Signed-JWT
    auth_url = "https://www.googleapis.com/oauth2/v4/token"

    issued = int(time.time())
    expires = issued + EXPIRE  # expires_in is in seconds

    # Note: this token expires and cannot be refreshed. The token must be recreated

    # JWT Headers
    additional_headers = {
        'kid': pkey_id,
        "alg": "RS256",  # Google uses SHA256withRSA
        "typ": "JWT"
    }

    # JWT Payload
    payload = {
        "iss": email,  # Issuer claim
        "sub": email,  # Issuer claim
        "aud": auth_url,  # Audience claim
        "iat": issued,  # Issued At claim
        "exp": expires,  # Expire time
        "scope": scope  # Permissions
    }

    # Encode the headers and payload and sign creating a Signed JWT (JWS)
    sig = jwt.encode(payload, pkey, algorithm="RS256", headers=additional_headers)
    return sig


def exchange_jwt_for_access_token(signed_jwt):
    """
    This function takes a Signed JWT and exchanges it for a Google OAuth Access Token
    """

    auth_url = "https://www.googleapis.com/oauth2/v4/token"

    params = {
        "grant_type": "urn:ietf:params:oauth:grant-type:jwt-bearer",
        "assertion": signed_jwt
    }

    r = requests.post(auth_url, data=params)

    if r.ok:
        return (r.json()['access_token'], '')
    return None, r.text


class GoogleIntegration(Integration):
    def __init__(self):
        Integration.__init__(self)

    def get_integration_config(self) -> str:
        try:
            pubsubs = ""
            module = GOOGLE_PUBSUB
            groups = get_module_group("GCP")
            if groups is not None and len(groups) > 0:
                for group in groups:
                    pubsub_configs = self.get_input_integration(module, 'jsonKey', group)
                    if pubsub_configs is not None:
                        json_filename = get_google_key_path(
                            self.decryptJsonKey(pubsub_configs["jsonKey"]),
                            pubsub_configs["projectId"])
                        cred = load_json_credentials(json_filename)

                        private_key = load_private_key(cred)

                        s_jwt = create_signed_jwt(
                            private_key,
                            cred['private_key_id'],
                            cred['client_email'],
                            SCOPE)

                        token, err = exchange_jwt_for_access_token(s_jwt)

                        if token is None:
                            print("Unable to start google pubsub: " + str(err))
                            pass

                        pubsubs += """
    google_pubsub {{
        project_id => "{}"
        id => "{}"
        add_field => {{ "[@metadata][dataSource]" => "{}" }}
        type => "gcp"
        topic => "{}"
        subscription => "{}"
        json_key_file => "{}"
    }}
                                    """.format(pubsub_configs["projectId"],group,group,pubsub_configs["topic"],pubsub_configs["subscription"],json_filename)

                return pubsubs
        except Exception as exception:
            print("Unable to start google pubsub: " + str(exception))

    def decryptJsonKey(self, jsonKey):
        jsonValue = AESCipher().decrypt(jsonKey)
        return json.loads(jsonValue)
