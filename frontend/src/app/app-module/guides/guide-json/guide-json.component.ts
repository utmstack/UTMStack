import {AfterViewInit, Component, OnInit} from '@angular/core';
import {FederationConnectionService} from '../../../app-management/connection-key/shared/services/federation-connection.service';

@Component({
  selector: 'app-guide-json',
  templateUrl: './guide-json.component.html',
  styleUrls: ['./guide-json.component.scss']
})
export class GuideJsonComponent implements OnInit, AfterViewInit {
  instance = window.location.host;
  mask = '**********************************';
  pythonCode = `
# Send logs using Python requests library
import requests
import json

# Define the URL
url = "https://${this.instance}:8080/v1/logs"

# Define the payload (replace this with your actual log data)
payload = json.dumps({
  "dataType": "json-input",
  "dataSource": "api-gateway",
  "raw": "{\\"level\\":\\"error\\",\\"msg\\":\\"connection refused\\",\\"user\\":\\"admin\\",\\"ip\\":\\"10.0.0.5\\",\\"endpoint\\":\\"/api/login\\",\\"status\\":500}"
})

# Define headers
headers = {
  'Utm-Connection-Key': '${this.mask}',
  'Content-Type': 'application/json'
}

# Make a POST request
response = requests.request("POST", url, headers=headers, data=payload)

# Print the response
print(response.text)
`;

  javaCode = `
// Create an OkHttpClient instance
OkHttpClient client = new OkHttpClient().newBuilder().build();

// Define the media type and request body (replace this with your actual request body)
MediaType mediaType = MediaType.parse("application/json");
RequestBody body = RequestBody.create(mediaType, "{\\"dataType\\":\\"json-input\\",\\"dataSource\\":\\"api-gateway\\",\\"raw\\":\\"{\\\\\\"level\\\\\\":\\\\\\"error\\\\\\",\\\\\\"msg\\\\\\":\\\\\\"connection refused\\\\\\",\\\\\\"user\\\\\\":\\\\\\"admin\\\\\\",\\\\\\"ip\\\\\\":\\\\\\"10.0.0.5\\\\\\",\\\\\\"endpoint\\\\\\":\\\\\\"/api/login\\\\\\",\\\\\\"status\\\\\\":500}\\"}");

// Create the request
Request request = new Request.Builder()
  .url("https://${this.instance}:8080/v1/logs")
  .method("POST", body)
  .addHeader("Utm-Connection-Key", "${this.mask}")
  .addHeader("Content-Type", "application/json")
  .build();

// Execute the request
Response response = client.newCall(request).execute();
`;

  golangCode = `
package main

import (
  "fmt"
  "strings"
  "net/http"
  "io/ioutil"
)

func main() {
  // Define the URL
  url := "https://${this.instance}:8080/v1/logs"
  method := "POST"

  // Define the payload (replace this with your actual payload)
  payload := strings.NewReader("{\\"dataType\\":\\"json-input\\",\\"dataSource\\":\\"api-gateway\\",\\"raw\\":\\"{\\\\\\"level\\\\\\":\\\\\\"error\\\\\\",\\\\\\"msg\\\\\\":\\\\\\"connection refused\\\\\\",\\\\\\"user\\\\\\":\\\\\\"admin\\\\\\",\\\\\\"ip\\\\\\":\\\\\\"10.0.0.5\\\\\\",\\\\\\"endpoint\\\\\\":\\\\\\"/api/login\\\\\\",\\\\\\"status\\\\\\":500}\\"}")

  // Create an HTTP client
  client := &http.Client{}

  // Create an HTTP request
  req, err := http.NewRequest(method, url, payload)

  if err != nil {
    fmt.Println(err)
    return
  }

  // Add headers
  req.Header.Add("Utm-Connection-Key", "${this.mask}")
  req.Header.Add("Content-Type", "application/json")

  // Send the request
  res, err := client.Do(req)

  if err != nil {
    fmt.Println(err)
    return
  }

  defer res.Body.Close()

  // Read and print the response
  body, err := ioutil.ReadAll(res.Body)

  if err != nil {
    fmt.Println(err)
    return
  }

  fmt.Println(string(body))
}
`;

  nodeJsCode = `
var request = require('request');

// Define the request options
var options = {
  'method': 'POST',
  'url': 'https://${this.instance}:8080/v1/logs',
  'headers': {
    'Utm-Connection-Key': '${this.mask}',
    'Content-Type': 'application/json'
  },
  // Define the request body (replace this with your actual request body)
  body: JSON.stringify({
    "dataType": "json-input",
    "dataSource": "api-gateway",
    "raw": "{\\"level\\":\\"error\\",\\"msg\\":\\"connection refused\\",\\"user\\":\\"admin\\",\\"ip\\":\\"10.0.0.5\\",\\"endpoint\\":\\"/api/login\\",\\"status\\":500}"
  })
};

// Send the request
request(options, function (error, response) {
  if (error) throw new Error(error);
  console.log(response.body);
});
`;

  powershellCode = `
# Define headers
$headers = New-Object "System.Collections.Generic.Dictionary[[String],[String]]"
$headers.Add("Utm-Connection-Key", "${this.mask}")
$headers.Add("Content-Type", "application/json")

# Define the request body (replace this with your actual request body)
$body = @"
{
  "dataType": "json-input",
  "dataSource": "api-gateway",
  "raw": "{\\"level\\":\\"error\\",\\"msg\\":\\"connection refused\\",\\"user\\":\\"admin\\",\\"ip\\":\\"10.0.0.5\\",\\"endpoint\\":\\"/api/login\\",\\"status\\":500}"
}
"@

# Send the request and capture the response
$response = Invoke-RestMethod 'https://${this.instance}:8080/v1/logs' -Method 'POST' -Headers $headers -Body $body

# Convert the response to JSON and print it
$response | ConvertTo-Json
`;

  curlCode = `
  curl --location --request POST 'https://${this.instance}:8080/v1/logs' \\
  --header 'Utm-Connection-Key: ${this.mask}' \\
  --header 'Content-Type: application/json' \\
  --data '{
    "dataType": "json-input",
    "dataSource": "api-gateway",
    "raw": "{\\"level\\":\\"error\\",\\"msg\\":\\"connection refused\\",\\"user\\":\\"admin\\",\\"ip\\":\\"10.0.0.5\\",\\"endpoint\\":\\"/api/login\\",\\"status\\":500}"
  }'
  `;
  token: string;
  view: 'python' | 'java' | 'javascript' | 'golang' | 'bash' | 'powershell' = 'python';
  exampleCode = `
  {
    "dataType": "json-input",
    "dataSource": "api-gateway",
    "raw": "{
      \\"level\\": \\"error\\",
      \\"msg\\": \\"connection refused\\",
      \\"user\\": \\"admin\\",
      \\"ip\\": \\"10.0.0.5\\",
      \\"endpoint\\": \\"/api/login\\",
      \\"status\\": 500
    }"
  }
  `;
  responseBody = `{"status": "received"}`;

  constructor(private federationConnectionService: FederationConnectionService) {
    this.getToken();
  }

  ngOnInit() {
  }

  getToken() {
    this.federationConnectionService.getToken().subscribe(response => {
      if (response.body !== null && response.body !== '') {
        this.token = response.body;
      } else {
        this.token = '';
      }

    });
  }


  ngAfterViewInit() {
  }


}
