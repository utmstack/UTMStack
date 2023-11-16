package com.utmstack.webtopdf.service;

import lombok.extern.slf4j.Slf4j;
import org.openqa.selenium.*;
import org.openqa.selenium.print.PageMargin;
import org.openqa.selenium.print.PageSize;
import org.openqa.selenium.print.PrintOptions;
import org.springframework.http.HttpHeaders;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Service;

import java.net.URL;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.concurrent.TimeUnit;

@Service
@Slf4j
public class PdfGenerationService {

    final private PrintOptions printOptions;

    private final WebDriver chromeDriver;

    public PdfGenerationService(WebDriver chromeDriver) {

        printOptions = new PrintOptions();
        printOptions.setPageMargin(new PageMargin(0, 0, 0, 0));
        printOptions.setPageSize(new PageSize(29.7, 21));

        this.chromeDriver = chromeDriver;
    }

    public byte[] generatePdf(String url) throws Exception {

        try {

            chromeDriver.get(url);

            TimeUnit.SECONDS.sleep(10);

            Pdf print = ((PrintsPage) chromeDriver).print(printOptions);

            return OutputType.BYTES.convertFromBase64Png(print.getContent());

        } catch (Exception e) {
            log.error(e.getMessage());
            chromeDriver.quit();
        }
        return new byte[0];
    }
}

