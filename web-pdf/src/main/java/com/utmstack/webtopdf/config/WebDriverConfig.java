package com.utmstack.webtopdf.config;

import lombok.extern.slf4j.Slf4j;
import org.openqa.selenium.WebDriver;
import org.openqa.selenium.chrome.ChromeOptions;
import org.openqa.selenium.remote.RemoteWebDriver;
import org.springframework.context.annotation.Configuration;

import java.net.MalformedURLException;
import java.net.URL;

@Configuration
@Slf4j
public class WebDriverConfig {

    String webDriverHost = System.getenv("WEB_DRIVER_HOST");
    String webDriverPort = System.getenv("WEB_DRIVER_PORT");

    public WebDriver createWebDriver() {
        try {

            URL serverUrl = new URL("http://" + webDriverHost + ":" + webDriverPort + "/wd/hub");

            ChromeOptions options = new ChromeOptions();
            options.addArguments("--headless");
            options.addArguments("--no-sandbox");

            return new RemoteWebDriver(serverUrl, options);
        } catch (MalformedURLException | RuntimeException exception) {
            log.error(exception.getMessage());
            throw new RuntimeException("Failed to initialize RemoteWebDriver", exception);
        }
    }
}