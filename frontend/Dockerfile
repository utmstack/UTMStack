FROM nginx:1.19.5
WORKDIR /usr/share/nginx/html
COPY ./nginx/mime.types /etc/nginx/mime.types
COPY ./dist/utm-stack/ /usr/share/nginx/html/

HEALTHCHECK --start-period=120s --interval=60s --timeout=60s \
  CMD curl --insecure -I -f http://localhost/health && curl --insecure -I -f http://localhost/api/ping || exit 1

EXPOSE 80

CMD ["/bin/sh",  "-c", "exec nginx -g 'daemon off;'"]