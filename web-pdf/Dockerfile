# Use a Selenium base image.
FROM selenium/standalone-chrome:latest

ADD target/web-pdf-1.0.0.jar ./

EXPOSE 8080

CMD /opt/bin/entry_point.sh & java -jar web-pdf-1.0.0.jar