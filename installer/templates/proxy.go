package templates

const ProxyUbuntu string = `map $http_upgrade $connection_upgrade {
    default upgrade;
    ''      close;
}

server {
    listen 80 default_server;
    server_name _;
    return 301 https://$host$request_uri;
}

server {
    # http2 so gRPC survives the hop. The old syntax rather than the http2
    # directive: this nginx comes from the distro, and Ubuntu and RHEL both
    # still ship versions that predate it.
    listen 443 ssl http2;
    server_name _;

    set $utmstack http://127.0.0.1:10001;
    set $shared_key {{.SharedKey}};

    # Agents speak gRPC, and proxy_pass below would answer them over HTTP/1.1,
    # which cannot carry it. Handed to the router in one piece: every service
    # is in the proto package "agent", and which one it is gets decided there.
    location /agent. {
        grpc_pass grpc://127.0.0.1:10001;
        grpc_set_header x-shared-key $shared_key;
        # Long-lived agent streams: nginx's gRPC keepalive PINGs are HTTP/2
        # control frames and do NOT reset these timers; only DATA does.
        # An AgentStream is pure DATA between commands, so the 60s-default
        # client_body_timeout (request body inactivity) is what evicted idle
        # agents at exactly 60s. Use a 24h inactivity backstop; liveness is
        # the app's job (gRPC keepalive + TCP keepalive).
        grpc_read_timeout 86400;
        grpc_send_timeout 86400;
        client_body_timeout 86400;
    }

    # log-input's ingest. Separate only because its proto package differs; the
    # router behind decides which of the two it is.
    location /plugins. {
        grpc_pass grpc://127.0.0.1:10001;
        grpc_set_header x-shared-key $shared_key;
        # Same long-lived-stream treatment: a quiet collector can easily go
        # more than 60s without a single log line on the stream.
        grpc_read_timeout 86400;
        grpc_send_timeout 86400;
        client_body_timeout 86400;
    }

    location / {
        proxy_pass  $utmstack;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $connection_upgrade;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header x-shared-key $shared_key;
        add_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_read_timeout 900;
    }

    error_page 502 /custom_502.html;
    location = /custom_502.html {
        root /etc/nginx/html;
        internal;
    }

    ssl_certificate /utmstack/cert/utm.crt;
    ssl_certificate_key /utmstack/cert/utm.key;
    ssl_protocols TLSv1.3;
    ssl_ciphers 'EECDH+AESGCM:EDH+AESGCM:AES256+EECDH:AES256+EDH';
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    chunked_transfer_encoding on;

    client_max_body_size 200M;
    client_body_buffer_size 200M;
}`

const ProxyRHEL string = `worker_processes  auto;

events {
    worker_connections  2048;
}

http {
    include       mime.types;
    default_type  application/octet-stream;

    keepalive_timeout  65;

    map $http_upgrade $connection_upgrade {
        default upgrade;
        ''      close;
    }

server {
    listen 80 default_server;
    server_name _;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name _;

    set $utmstack http://127.0.0.1:10001;
    set $shared_key {{.SharedKey}};

    location /agent. {
        grpc_pass grpc://127.0.0.1:10001;
        grpc_set_header x-shared-key $shared_key;
        # Long-lived agent streams: nginx's gRPC keepalive PINGs are HTTP/2
        # control frames and do NOT reset these timers; only DATA does.
        # An AgentStream is pure DATA between commands, so the 60s-default
        # client_body_timeout (request body inactivity) is what evicted idle
        # agents at exactly 60s. Use a 24h inactivity backstop; liveness is
        # the app's job (gRPC keepalive + TCP keepalive).
        grpc_read_timeout 86400;
        grpc_send_timeout 86400;
        client_body_timeout 86400;
    }

    # log-input's ingest. Separate only because its proto package differs; the
    # router behind decides which of the two it is.
    location /plugins. {
        grpc_pass grpc://127.0.0.1:10001;
        grpc_set_header x-shared-key $shared_key;
        # Same long-lived-stream treatment: a quiet collector can easily go
        # more than 60s without a single log line on the stream.
        grpc_read_timeout 86400;
        grpc_send_timeout 86400;
        client_body_timeout 86400;
    }

    location / {
        proxy_pass  $utmstack;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $connection_upgrade;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header x-shared-key $shared_key;
        add_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_read_timeout 900;
    }

    error_page 502 /custom_502.html;
    location = /custom_502.html {
        root /etc/nginx/html;
        internal;
    }

    ssl_certificate /utmstack/cert/utm.crt;
    ssl_certificate_key /utmstack/cert/utm.key;
    ssl_protocols TLSv1.3;
    ssl_ciphers 'EECDH+AESGCM:EDH+AESGCM:AES256+EECDH:AES256+EDH';
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    chunked_transfer_encoding on;

    client_max_body_size 200M;
    client_body_buffer_size 200M;
   }
}`
