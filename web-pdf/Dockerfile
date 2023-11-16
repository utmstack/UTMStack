FROM openjdk:17-jdk-alpine

RUN apk add --no-cache curl
RUN apk update && \
    apk add chromium chromium-chromedriver

WORKDIR /app

RUN mkdir webpdf

COPY target/webToPdf-.1.0.0.jar app.jar

EXPOSE 8080

CMD ["java", "-jar", "app.jar"]