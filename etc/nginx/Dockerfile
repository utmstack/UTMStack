FROM nginx:1.19.5
COPY panel.conf /etc/nginx/conf.d/default.conf

HEALTHCHECK --start-period=120s --interval=60s --timeout=60s \
  CMD curl --insecure -I -f https://localhost/health && curl --insecure -I -f https://localhost/utmstack/api/ping || exit 1
