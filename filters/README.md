# Logstash filters used by UTMStack
This part of the project, represents the default system filters used by logstash to proccess logs data coming from many sources.

## Getting Started
Before you begin, you must know the basics about git and Logstash. Logstash documentation can be found at: https://www.elastic.co/guide/en/logstash/current/getting-started-with-logstash.html

1. **Clone the Repository**: If not already done, clone the repository to your local machine:
   ```bash
   cd [your-repository-dir]
   git clone https://github.com/utmstack/UTMStack.git
    ```

2. **Download logstash**: Download 7.14.x logstash OSS version for Windows from: https://www.elastic.co/downloads/logstash-oss, and extract files to a folder.

3. **Configure logstash**: To test a filter you have to create a file with .conf extension, for example: myfilter.conf. Then, inside that file, you have to create an input, this input can be: tcp or file, see the examples below.

```
input {
  tcp{
    port=> 514
    type=> "syslog"
  }
}
or
input {
  file {
    path => "your_path_to_log_file/log_file.log"
	sincedb_path => "nul"
	start_position => beginning
  }
}
```
If the filter you're testing have the `split` plugin at the top of the code, then the input log lines must be in single line separated by the string defined in the `terminator` value of the plugin. See fortinet log line for example:

`<123>devname=firewall device_id= date=2023-08-01 time=22:45:56 logid="0000000013" type="traffic" subtype="forward" level="notice" vd="root" eventtime=1629312357382231219 tz="+0400" srcip=10.10.10.2 srcname="" srcport=52150 srcintf="vlan" srcintfrole="lan" dstip=10.10.10.2 dstport=53 dstintf="lan" dstintfrole="lan" sessionid=26383724 proto=17 action="accept" policyid=23 policytype="policy" service="DNS" dstcountry="Reserved" srccountry="Reserved" trandisp="noop" appid=16195 app="DNS" appcat="Network.Service" apprisk="elevated" applist="default" appact="detected" duration=181 sentbyte=74 rcvdbyte=199 sentpkt=1 rcvdpkt=1 utmaction="allow" countdns=1 osname="Windows" srcswversion="10" unauthuser="" unauthusersource="kerberos" mastersrcmac="xx:xx:xx:xx:xx:xx" srcmac="xx:xx:xx:xx:xx:xx" srcserver=0<utm-log-separator><189>date=2023-01-17 time=17:35:53 devname="myhost" devid="FGxxxxxxxxxxx" logid="0000000013" type="traffic" subtype="forward" level="notice" vd="root" eventtime=1642440953 srcip=10.10.10.2 srcport=46824 srcintf="wan2" srcintfrole="undefined" dstip=10.10.10.3 dstport=1723 dstintf="port1" dstintfrole="undefined" poluuid="xxxxxxxxx-5451-51ec-c5c3-xxxxxxx" sessionid=7108610 proto=6 action="timeout" policyid=48 policytype="policy" service="PPTP" dstcountry="Reserved" srccountry="United States" trandisp="dnat" tranip=10.10.10.2 tranport=1723 duration=30 sentbyte=40 rcvdbyte=68 sentpkt=1 rcvdpkt=1 appcat="unscanned" crscore=5 craction=262144 crlevel="low"<utm-log-separator>`

If you need to test the input logs in separated lines, you must remove the split plugin from filter and the separator from the input log lines as follows:
```
<123>devname=firewall device_id= date=2023-08-01 time=22:45:56 logid="0000000013" type="traffic" subtype="forward" level="notice" vd="root" eventtime=1629312357382231219 tz="+0400" srcip=10.10.10.2 srcname="" srcport=52150 srcintf="vlan" srcintfrole="lan" dstip=10.10.10.2 dstport=53 dstintf="lan" dstintfrole="lan" sessionid=26383724 proto=17 action="accept" policyid=23 policytype="policy" service="DNS" dstcountry="Reserved" srccountry="Reserved" trandisp="noop" appid=16195 app="DNS" appcat="Network.Service" apprisk="elevated" applist="default" appact="detected" duration=181 sentbyte=74 rcvdbyte=199 sentpkt=1 rcvdpkt=1 utmaction="allow" countdns=1 osname="Windows" srcswversion="10" unauthuser="" unauthusersource="kerberos" mastersrcmac="xx:xx:xx:xx:xx:xx" srcmac="xx:xx:xx:xx:xx:xx" srcserver=0`
<189>date=2023-01-17 time=17:35:53 devname="myhost" devid="FGxxxxxxxxxxx" logid="0000000013" type="traffic" subtype="forward" level="notice" vd="root" eventtime=1642440953 srcip=10.10.10.2 srcport=46824 srcintf="wanx" srcintfrole="undefined" dstip=10.10.10.3 dstport=1723 dstintf="port1" dstintfrole="undefined" poluuid="xxxxxxxx-5451-51ec-c5c3-xxxxxxx" sessionid=7108610 proto=6 action="timeout" policyid=48 policytype="policy" service="PPTP" dstcountry="Reserved" srccountry="Reserved" trandisp="dnat" tranip=10.10.10.2 tranport=1723 duration=30 sentbyte=40 rcvdbyte=68 sentpkt=1 rcvdpkt=1 appcat="unscanned" crscore=5 craction=262144 crlevel="low"
```

If you are sending the logs by tcp, you must follow the same principle, one line separated by `terminator` value if split is present or individual lines without the separator if not.

Then, copy the content of the filter to test to your file after the input definition, remember the split usage mentioned before. Configure the output to a file, copy the output configuration after the filter section, for example:

```
output {
    file {
    path => "your_output_path/out.log"
  }
}
```

Your filter must be similar to this structure
```
input {
  ...
}
filter{
  ...  
}
output{
  ...
}
```

4. **Test your filter**: Once you have your filter file configured, you have to go to your logstash installation bin folder and execute:
  - Windows example: `D:\logstash-7.14.1\bin>logstash -r -f path_to_your_filter\myfilter.conf`
  - Docker example: Will be added later

