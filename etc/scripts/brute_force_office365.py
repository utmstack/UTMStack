
import requests
import random
import string

url = 'https://login.microsoftonline.com/organizations/oauth2/v2.0/token'

# Try to login with invalid credentials 10 times
for i in range(10):
    invalid_password = ''.join(random.choices(string.ascii_lowercase + string.digits, k=10))
    payload = {
        'grant_type': 'password',
        'client_id': '', # must be a valid client_id 
        'client_secret': '', # must be a valid client_secret
        'scope': 'https://graph.microsoft.com/.default',
        'username': '', # must be a valid username 
        'password': invalid_password
    }

    
    response = requests.post(url, data=payload)

    print(f"Attempt {i+1}: {response.status_code}")
    print(response.json())

# Login with valid credentials
payload = {
    'grant_type': 'password',
    'client_id': '', # must be a valid client_id 
    'client_secret': '', # must be a valid client_secret
    'scope': 'https://graph.microsoft.com/.default',
    'username': '', # must be a valid username 
    'password': '' # must be a valid password
}

response = requests.post(url, data=payload)

print(f"Valid credentials: {response.status_code}")
print(response.json())
