FROM nginx:1.31.8-alpine

WORKDIR /usr/share/nginx/html
RUN rm -rf ./*

COPY ./dist/ .
COPY ./nginx/default.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
