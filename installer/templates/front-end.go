package templates

const FrontEnd string = `server {
    listen 80;
    server_name _;
    resolver 127.0.0.11 valid=10s ipv6=off;

    location / {
      root /usr/share/nginx/html;
      index index.html index.htm;
      try_files $uri $uri/ /index.html =404;

      # The SPA renders SIEM data; these bound what an injected payload can do
      # and stop the panel from being framed by another site.
      add_header X-Frame-Options "SAMEORIGIN" always;
      add_header X-Content-Type-Options "nosniff" always;
      add_header Referrer-Policy "strict-origin-when-cross-origin" always;
      add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' data: https://fonts.gstatic.com; img-src 'self' data: blob:; connect-src 'self' ws: wss:; frame-ancestors 'self'; form-action 'self'; base-uri 'self'; object-src 'none'" always;
      # Reported, not enforced: index.html still carries one inline script and
      # inline styles. Once those are moved out (or hashed), fold script-src
      # and style-src below into the enforced policy above.
      add_header Content-Security-Policy-Report-Only "default-src 'self'; script-src 'self'; style-src 'self' https://fonts.googleapis.com; object-src 'none'" always;
    }

    set $utmstack_backend http://backend:8080;
    set $utmstack_agent_manager http://agentmanager:9001;
    set $utmstack_backend_auth http://backend:8080/api/authenticate;
    set $utmstack_ws http://backend:8080/ws;
    set $shared_key {{.SharedKey}};
    set $shared_key_header $http_x_shared_key;

	location /saml2/ {
    	proxy_pass http://backend:8080/saml2/;
    	proxy_set_header Host $host;
    	proxy_set_header X-Real-IP $remote_addr;
    	proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    	proxy_set_header X-Forwarded-Proto $scheme;
	}

	location /login/saml2/ {
    	proxy_pass http://backend:8080/login/saml2/;
    	proxy_set_header Host $host;
    	proxy_set_header X-Real-IP $remote_addr;
    	proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    	proxy_set_header X-Forwarded-Proto $scheme;
	}

    location /api {
        proxy_pass  $utmstack_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 900;
        if ($shared_key_header != $shared_key){
             return 403;
        }
    }

    location /api/ping {
        proxy_pass  $utmstack_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 900;
    }

    location /ws {
        proxy_pass $utmstack_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "Upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }

   location /management {
        proxy_pass  $utmstack_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 900;
    }

    location /swagger-ui {
        proxy_pass  $utmstack_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 900;
    }

    location /dependencies {
        proxy_pass  $utmstack_agent_manager;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 900;
    }

    location /v3 {
        proxy_pass  $utmstack_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 900;
    }

    client_max_body_size 200M;
    client_body_buffer_size 200M;
}`
