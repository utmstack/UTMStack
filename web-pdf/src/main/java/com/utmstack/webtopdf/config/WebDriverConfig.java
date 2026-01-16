package com.utmstack.webtopdf.config;

import lombok.extern.slf4j.Slf4j;
import org.openqa.selenium.WebDriver;
import org.openqa.selenium.chrome.ChromeOptions;
import org.openqa.selenium.remote.RemoteWebDriver;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Configuration;

import java.net.MalformedURLException;
import java.net.URL;
import java.time.Duration;

@Configuration
@Slf4j
public class WebDriverConfig {

    @Value("${selenium.grid.url:http://localhost:4444/wd/hub}")
    private String seleniumGridUrl;

    public WebDriver createWebDriver() {
        try {
            WebDriver driver = getWebDriver();

            driver.manage().timeouts().pageLoadTimeout(Duration.ofSeconds(60));
            driver.manage().timeouts().scriptTimeout(Duration.ofSeconds(30));

            return driver;

        } catch (Exception exception) {
            log.error("Failed to initialize RemoteWebDriver", exception);
            throw new RuntimeException("Failed to initialize RemoteWebDriver", exception);
        }
    }

    private WebDriver getWebDriver() throws MalformedURLException {
        URL serverUrl = new URL(seleniumGridUrl);

        ChromeOptions options = new ChromeOptions();
        options.addArguments("--headless=new");
        options.addArguments("--no-sandbox");
        options.addArguments("--disable-dev-shm-usage");
        options.addArguments("--disable-gpu");
        options.addArguments("--disable-software-rasterizer");
        options.addArguments("--window-size=1920,1080");
        options.addArguments("--remote-allow-origins=*");
        options.setAcceptInsecureCerts(true);

        return new RemoteWebDriver(serverUrl, options);
    }
}
