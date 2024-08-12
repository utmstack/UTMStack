package com.utmstack.webtopdf.service;

import com.utmstack.webtopdf.config.WebDriverConfig;
import com.utmstack.webtopdf.config.enums.AccessType;
import lombok.extern.slf4j.Slf4j;
import org.openqa.selenium.*;
import org.openqa.selenium.print.PageMargin;
import org.openqa.selenium.print.PageSize;
import org.openqa.selenium.print.PrintOptions;
import org.openqa.selenium.support.ui.ExpectedConditions;
import org.openqa.selenium.support.ui.WebDriverWait;
import org.springframework.stereotype.Service;

import java.time.Duration;

@Service
@Slf4j
public class PdfGenerationService {

    final private PrintOptions printOptions;

    private final WebDriverConfig webDriverConfig;

    public PdfGenerationService(WebDriverConfig webDriverConfig) {

        printOptions = new PrintOptions();
        printOptions.setPageMargin(new PageMargin(0, 0, 0, 0));
        printOptions.setPageSize(new PageSize(29.7, 21));

        this.webDriverConfig = webDriverConfig;
    }

    public byte[] generatePdf(String url, String route, String accessKey, AccessType accessType) {

        WebDriver webDriver = webDriverConfig.createWebDriver();

        try {
            String report = url.concat(accessType.buildUrlPart(accessKey, route));
            webDriver.get(report);

            WebDriverWait wait = new WebDriverWait(webDriver, Duration.ofSeconds(30));
            wait.until(ExpectedConditions.presenceOfElementLocated(By.id("report")));

            Pdf print = ((PrintsPage) webDriver).print(printOptions);
            webDriver.quit();

            return OutputType.BYTES.convertFromBase64Png(print.getContent());

        } catch (Exception e) {
            log.error(e.getMessage());
            webDriver.quit();
        }
        return new byte[0];
    }
}