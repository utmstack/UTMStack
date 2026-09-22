#!/usr/bin/env bash
# remove_utm.bash — purga total de UTMStack de la máquina.
# Idempotente: se puede correr varias veces sin romper nada.
# Requiere: sudo. NO desinstala Docker ni el SO.

set -u

if [[ $EUID -ne 0 ]]; then
    echo "run as root: sudo $0"
    exit 1
fi

log() { printf "\n\033[1;36m==> %s\033[0m\n" "$*"; }
ok()  { printf "    \033[1;32m[OK]\033[0m %s\n" "$*"; }

log "1/8 stopping UTMStackComponentsUpdater service"
systemctl stop UTMStackComponentsUpdater 2>/dev/null || true
systemctl disable UTMStackComponentsUpdater 2>/dev/null || true
rm -f /etc/systemd/system/UTMStackComponentsUpdater.service
systemctl daemon-reload
systemctl reset-failed 2>/dev/null || true
ok "service removed"

log "2/8 removing docker stack"
docker stack rm utmstack 2>/dev/null || true
sleep 30
docker service ls -q 2>/dev/null | xargs -r docker service rm 2>/dev/null || true
ok "stack removed"

log "3/8 removing all containers, volumes, networks, images, build cache"
docker ps -aq 2>/dev/null | xargs -r docker rm -f 2>/dev/null || true
docker volume ls -q 2>/dev/null | xargs -r docker volume rm -f 2>/dev/null || true
docker network ls -q --filter type=custom 2>/dev/null | xargs -r docker network rm 2>/dev/null || true
docker system prune -a -f --volumes 2>/dev/null || true
ok "docker purged"

log "4/8 leaving swarm"
docker swarm leave --force 2>/dev/null || true
ok "swarm left"

log "5/8 removing /utmstack data + installer state"
rm -rf /utmstack
rm -f /root/utmstack.yml
rm -f /root/step.txt
ok "data removed"

log "6/8 removing host nginx (config + package)"
systemctl stop nginx 2>/dev/null || true
rm -f /etc/nginx/sites-available/default
rm -f /etc/nginx/sites-enabled/default
rm -f /etc/nginx/nginx.conf
rm -f /etc/nginx/html/custom_502.html
if command -v apt-get >/dev/null 2>&1; then
    DEBIAN_FRONTEND=noninteractive apt-get purge -y nginx nginx-common nginx-core 2>/dev/null || true
elif command -v dnf >/dev/null 2>&1; then
    dnf remove -y nginx 2>/dev/null || true
fi
rm -rf /etc/nginx /var/log/nginx
ok "nginx removed"

log "7/8 removing installer logs"
rm -rf /var/log/utmstack
ok "logs removed"

log "8/8 verification"
echo "-- docker containers:"
docker ps -a 2>/dev/null || echo "(docker not available)"
echo "-- docker volumes:"
docker volume ls 2>/dev/null || true
echo "-- utmstack files/dirs still on disk:"
ls -la /utmstack /root/utmstack.yml /root/step.txt 2>&1 | grep -v "No such" || echo "(none)"
echo "-- systemd units referencing utm:"
systemctl list-units --all 2>/dev/null | grep -i utm || echo "(none)"

printf "\n\033[1;32mDONE.\033[0m Ready for a clean install.\n"
