package templates

const Proxy string = `server {
    listen 80 default_server;
    server_name _;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl;
    server_name _;

    set $utmstack http://127.0.0.1:10001;
    set $shared_key {{.SharedKey}};

    location / {
        proxy_pass  $utmstack;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header x-shared-key $shared_key;
        add_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_read_timeout 900;
    }

    ssl_certificate /utmstack/cert/utm.crt;
    ssl_certificate_key /utmstack/cert/utm.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers 'EECDH+AESGCM:EDH+AESGCM:AES256+EECDH:AES256+EDH';
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    chunked_transfer_encoding on;

    client_max_body_size 200M;
    client_body_buffer_size 200M;
}`