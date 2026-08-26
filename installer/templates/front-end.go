package templates

const FrontEnd string = `map $http_upgrade $connection_upgrade {
    default upgrade;
    ''      close;
}

server {
    listen 80;
    http2 on;
    server_name _;
    resolver 127.0.0.11 valid=10s ipv6=off;

    set $utmstack_backend http://backend:8080;
    # https, not http: the dependency server terminates TLS itself with the
    # stack's certificate, which is the same one this proxy is fronted with.
    set $utmstack_agent_manager https://agentmanager:9001;
    set $utmstack_log_input http://log-input:50052;
    set $utmstack_agent_manager_grpc agentmanager:9000;
    set $utmstack_log_input_grpc log-input:50051;
    set $shared_key {{.SharedKey}};
    set $shared_key_header $http_x_shared_key;

    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection $connection_upgrade;
    proxy_http_version 1.1;
    proxy_read_timeout 900;

    location / {
      root /usr/share/nginx/html;
      index index.html index.htm;
      try_files $uri /index.html =404;
    }

    location /api {
        proxy_pass $utmstack_backend;
        if ($shared_key_header != $shared_key){
             return 403;
        }
    }

    location /api/ping {
        rewrite ^/api/ping$ /api/v1/health break;
        proxy_pass $utmstack_backend;
    }

    location /uploads {
        proxy_pass $utmstack_backend;
    }

    location /swagger {
        proxy_pass $utmstack_backend;
    }

    # The path the agent-manager actually serves. Installers, agents and
    # collectors all fetch binaries here.
    location /private/dependencies {
        proxy_pass $utmstack_agent_manager;
        # A self-signed certificate on an internal name: the hop is inside the
        # overlay, and what authenticates it is not being reachable from outside.

    }

    location /v1/ingest {
        proxy_pass $utmstack_log_input;
        proxy_request_buffering off;
    }

    location /agent.AgentService/ {
        grpc_pass grpcs://$utmstack_agent_manager_grpc;
        grpc_ssl_verify off;
        grpc_read_timeout 900;
        grpc_send_timeout 900;
    }

    location /agent.PanelService/ {
        grpc_pass grpcs://$utmstack_agent_manager_grpc;
        grpc_ssl_verify off;
        grpc_read_timeout 900;
        grpc_send_timeout 900;
    }

    location /agent.CollectorService/ {
        grpc_pass grpcs://$utmstack_agent_manager_grpc;
        grpc_ssl_verify off;
        grpc_read_timeout 900;
        grpc_send_timeout 900;
    }

    # log-input's ingest, whose service lives in the SDK's "plugins" package.
    location /plugins.Integration/ {
        grpc_pass grpcs://$utmstack_log_input_grpc;
        grpc_ssl_verify off;
        grpc_read_timeout 900;
        grpc_send_timeout 900;
    }

    location /agent.PingService/ {
        grpc_pass grpcs://$utmstack_agent_manager_grpc;
        grpc_ssl_verify off;
        grpc_read_timeout 900;
        grpc_send_timeout 900;
    }

    client_max_body_size 200M;
    client_body_buffer_size 200M;
}`
