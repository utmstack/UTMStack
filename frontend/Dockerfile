FROM nginx:1.25.3-alpine

WORKDIR /usr/share/nginx/html
RUN rm -rf ./*

COPY ./dist/ .
COPY ./nginx/default.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
