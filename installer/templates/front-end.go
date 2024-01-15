package templates

const FrontEnd string =`server {
    listen 80;
    server_name _;
    resolver 127.0.0.11 valid=10s ipv6=off;

    location / {
      root /usr/share/nginx/html;
      index index.html index.htm;
      try_files $uri $uri/ /index.html =404;
      add_header Content-Security-Policy "default-src 'self' https://fonts.googleapis.com/css* https://fonts.gstatic.com/s/poppins/v20*; frame-src 'self' data:; script-src 'self' https://storage.googleapis.com; style-src 'self'; img-src 'self' data:;  font-src 'self' data:" always;
    }

    set $utmstack_backend http://backend:8080;
    set $utmstack_filebrowser http://filebrowser:9091;
    set $utmstack_backend_auth http://backend:8080/api/authenticate;
    set $utmstack_ws http://backend:8080/ws;
    set $shared_key {{.SharedKey}};
    set $shared_key_header $http_x_shared_key;

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

    location /health {
        proxy_pass $utmstack_filebrowser;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }

    location /kill {
        set $auth_token "Bearer $cookie_utmauth";
        proxy_pass $utmstack_backend_auth;
        proxy_http_version 1.1;
        proxy_pass_request_body off;
        proxy_set_header Authorization $auth_token;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
    }

    location /srv {
        auth_request /kill;
        proxy_pass $utmstack_filebrowser;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }

    location /swagger-ui {
        proxy_pass  $utmstack_backend;
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