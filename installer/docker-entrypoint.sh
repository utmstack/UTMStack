#!/usr/bin/env bash
set -euo pipefail

log() { echo ">>> [utmstack-entrypoint] $*"; }

# Docker daemon: run unprivileged-friendly, with iptables off when needed.
if [ ! -S /var/run/docker.sock ]; then
    log "starting dockerd"
    nohup dockerd \
        --host=unix:///var/run/docker.sock \
        --storage-driver=overlay2 \
        > /var/log/dockerd.log 2>&1 &
fi

log "waiting for dockerd"
for i in $(seq 1 60); do
    if docker info >/dev/null 2>&1; then
        log "dockerd ready"
        break
    fi
    sleep 1
done
if ! docker info >/dev/null 2>&1; then
    log "dockerd never came up — see /var/log/dockerd.log"
    tail -n 80 /var/log/dockerd.log || true
    exit 1
fi

# The installer's CheckIfServiceIsInstalled() goes through kardianos/service
# which calls systemctl. We don't run systemd in this image, so shim it.
if ! command -v systemctl >/dev/null 2>&1 || ! systemctl is-system-running >/dev/null 2>&1; then
    log "installing systemctl shim (no real systemd in container)"
    cat >/usr/local/sbin/systemctl <<'SHIM'
#!/usr/bin/env bash
# Minimal shim: return success for status/is-enabled, no-op for the rest.
case "${1:-}" in
    status|is-active|is-enabled)
        echo "inactive"
        exit 3 ;;
    *) exit 0 ;;
esac
SHIM
    chmod +x /usr/local/sbin/systemctl
    hash -r
fi

if [ -f /root/utmstack.yml ] && [ -f /utmstack/updates/version.json ]; then
    log "existing install detected — restarting stack"
    docker stack deploy -c /root/utmstack-stack.yml utmstack 2>/dev/null || true
else
    log "running installer"
    /usr/local/bin/utmstack_installer --install
fi

log "install finished — tailing dockerd log to keep container alive"
exec tail -f /var/log/dockerd.log
