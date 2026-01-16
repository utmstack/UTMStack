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

        String reportUrl = String.format("%s%s", url, accessType.buildUrlPart(accessKey, route));
        WebDriver webDriver = webDriverConfig.createWebDriver();

        try {
            webDriver.get(reportUrl);

            WebDriverWait wait = new WebDriverWait(webDriver, Duration.ofSeconds(15));

            wait.until(d -> ((JavascriptExecutor) d)
                    .executeScript("return document.readyState").equals("complete"));

            Thread.sleep(500);

            wait.until(ExpectedConditions.presenceOfElementLocated(By.cssSelector(".report-loading")));

            Pdf print = ((PrintsPage) webDriver).print(printOptions);
            return OutputType.BYTES.convertFromBase64Png(print.getContent());

        } catch (TimeoutException e) {
            log.error("Timeout waiting for report to load: {}", e.getMessage());
            throw new TimeoutException("The report took too long to load.");

        } catch (NoSuchElementException e) {
            log.error("Required element not found: {}", e.getMessage());
            throw new NoSuchElementException("A required element was not found while generating the PDF.");

        } catch (Exception e) {
            log.error("Unexpected error generating PDF: {}", e.getMessage(), e);
            throw new RuntimeException("Unexpected error generating the PDF.");

        } finally {
            try {
                webDriver.quit();
            } catch (Exception ex) {
                log.warn("Error closing WebDriver: {}", ex.getMessage());
            }
        }
    }

}