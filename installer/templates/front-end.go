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

    # The agent's persistent gRPC streams to agent-manager. nginx has no
    # "no timeout" for upstream read/send (0 means "fire immediately", which
    # breaks every request), so use a large inactivity backstop of 24h.
    # gRPC keepalive PINGs are HTTP/2 control frames - nginx only ACKs them
    # and they do NOT reset these timers; only DATA does. Every timer below
    # must therefore be large for streams legitimately silent between calls.
    location /agent.AgentService/ {
        grpc_pass grpcs://$utmstack_agent_manager_grpc;
        grpc_read_timeout 86400;
        grpc_send_timeout 86400;
        client_body_timeout 86400;
        grpc_socket_keepalive on;
    }

    location /agent.PanelService/ {
        grpc_pass grpcs://$utmstack_agent_manager_grpc;
        grpc_read_timeout 86400;
        grpc_send_timeout 86400;
        client_body_timeout 86400;
        grpc_socket_keepalive on;
    }

    location /agent.CollectorService/ {
        grpc_pass grpcs://$utmstack_agent_manager_grpc;
        grpc_read_timeout 86400;
        grpc_send_timeout 86400;
        client_body_timeout 86400;
        grpc_socket_keepalive on;
    }

    # log-input's ingest, whose service lives in the SDK's "plugins" package.
    # log-input's ingest: also a persistent collector stream, so the same
    # 24h inactivity backstop (a quiet host can go >60s without a log line).
    location /plugins.Integration/ {
        grpc_pass grpcs://$utmstack_log_input_grpc;
        grpc_read_timeout 86400;
        grpc_send_timeout 86400;
        client_body_timeout 86400;
    }

    location /agent.PingService/ {
        grpc_pass grpcs://$utmstack_agent_manager_grpc;
        grpc_read_timeout 86400;
        grpc_send_timeout 86400;
        client_body_timeout 86400;
        grpc_socket_keepalive on;
    }

    client_max_body_size 200M;
    client_body_buffer_size 200M;
}`
