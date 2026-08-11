package templates

const FrontEnd string = `server {
    listen 80;
    server_name _;
    resolver 127.0.0.11 valid=10s ipv6=off;

    location / {
      root /usr/share/nginx/html;
      index index.html index.htm;
      try_files $uri /index.html =404;
    }

    set $utmstack_backend http://backend:8080;
    set $utmstack_agent_manager http://agentmanager:9001;
    set $shared_key {{.SharedKey}};
    set $shared_key_header $http_x_shared_key;

    location ~ ^/api/v1/soar/ws/ {
        proxy_pass $utmstack_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "Upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 900;
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
        # Health probe without the shared-key gate; mapped to the Go backend health.
        rewrite ^/api/ping$ /api/v1/health break;
        proxy_pass  $utmstack_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 900;
    }

    location /uploads {
        # Backend-served static assets (avatars + white-label branding). No
        # shared-key gate: the login page must load branding while unauthenticated.
        proxy_pass  $utmstack_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 900;
    }


    location /swagger {
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


    client_max_body_size 200M;
    client_body_buffer_size 200M;
}`
