# UTMStack Datasource for BitdefenderCloud
### Version 1.0.0

## Decription

UTMStack Datasource for BitdefenderCloud is a connector developed in golang that receives logs from Bitdefender GravityZone Cloud and sends them to syslog server

The connector uses the POST method to receive authenticated and protected messages from the GravityZone event push service. Parses the message and then forwards it to a UTMStack Syslog server

 