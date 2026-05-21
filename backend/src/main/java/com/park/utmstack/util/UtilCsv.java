package com.park.utmstack.util;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.jayway.jsonpath.DocumentContext;
import com.jayway.jsonpath.JsonPath;
import com.jayway.jsonpath.PathNotFoundException;
import com.park.utmstack.domain.shared_types.DataColumn;
import com.park.utmstack.util.exceptions.UtmCsvException;
import io.jsonwebtoken.lang.Assert;
import org.apache.commons.csv.CSVFormat;
import org.apache.commons.csv.CSVPrinter;
import org.apache.commons.csv.QuoteMode;
import org.springframework.http.HttpHeaders;
import org.springframework.util.StringUtils;

import javax.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.time.Instant;
import java.time.format.DateTimeFormatter;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.Locale;
import java.util.Map;
import java.util.stream.Collectors;
import java.util.stream.Stream;

public class UtilCsv {
    private static final String CLASS_NAME = "UtilCsv";
    private static final ObjectMapper OBJECT_MAPPER = new ObjectMapper();

    /**
     * Build a csv file with columns and data and write it to HttpServletResponse writer
     *
     * @param response : Http servlet to write the csv file
     * @param columns  : Headers of csv file
     * @param data     : Rows of csv file
     * @throws UtmCsvException In case of any error
     */
    public static void prepareToDownload(HttpServletResponse response, DataColumn[] columns, List<?> data) throws
            UtmCsvException {
        final String ctx = CLASS_NAME + ".prepareToDownload";
        final DateTimeFormatter DATE_FORMATTER = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss z")
                .withLocale(Locale.getDefault()).withZone(TimezoneUtil.getAppTimezone());
        try {
            Assert.notEmpty(columns);
            Assert.notEmpty(data);

            // Cleaning column names from .keyword termination
            Arrays.stream(columns).forEach(column ->
                    column.setField(column.getField().replace(".keyword", "")));

            List<String[]> rows = new ArrayList<>();

            data.forEach(d -> {
                DocumentContext docctx = JsonPath.parse(d);
                String[] cells = new String[columns.length];
                for (int i = 0; i < columns.length; i++) {
                    String fieldName = columns[i].getField();
                    String fieldType = columns[i].getType();
                    cells[i] = null;

                    Object value;
                    try {
                        value = docctx.read("$." + fieldName);
                    } catch (PathNotFoundException e) {
                        continue;
                    }

                    if (value == null)
                        continue;

                    if (value instanceof String) {
                        cells[i] = fieldType.equals("date") ? DATE_FORMATTER.format(Instant.parse(String.valueOf(value))) :
                                String.valueOf(value).replace("\n", " ").replace("\t", " ");
                    } else if (value instanceof List) {
                        cells[i] = ((List<?>) value).stream().map(String::valueOf).collect(Collectors.joining(","));
                    } else if (value instanceof Number) {
                        cells[i] = String.valueOf(value);
                    } else if (value instanceof Map) {
                        try {
                            cells[i] = OBJECT_MAPPER.writeValueAsString(value);
                        } catch (Exception ex) {
                            cells[i] = value.toString();
                        }
                    }
                    cells[i] = sanitizeCsvCell(cells[i]);
                }
                rows.add(cells);
            });

            String[] headers = Stream.of(columns).map(column -> {
                if (StringUtils.hasText(column.getLabel()))
                    return column.getLabel();
                return column.getField().replace(".keyword", "");
            }).toArray(String[]::new);

            response.setContentType("text/csv");
            response.setHeader(HttpHeaders.CONTENT_DISPOSITION, "attachment; filename=data.csv");

            CSVPrinter csvPrinter = new CSVPrinter(response.getWriter(), CSVFormat.DEFAULT.withHeader(headers)
                    .withQuoteMode(QuoteMode.ALL));
            for (String[] row : rows)
                csvPrinter.printRecords((Object) row);

        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            throw new UtmCsvException(msg);
        }
    }

    /**
     * Opens a CSV response stream: sets content-type/disposition headers and writes the header row.
     * Caller is responsible for closing the returned printer (try-with-resources is fine).
     *
     * Column names are normalized in-place by stripping a trailing {@code .keyword}.
     */
    public static CSVPrinter openCsvStream(HttpServletResponse response, DataColumn[] columns) throws IOException {
        Assert.notEmpty(columns);

        Arrays.stream(columns).forEach(column ->
                column.setField(column.getField().replace(".keyword", "")));

        String[] headers = Stream.of(columns).map(column -> {
            if (StringUtils.hasText(column.getLabel()))
                return column.getLabel();
            return column.getField().replace(".keyword", "");
        }).toArray(String[]::new);

        response.setContentType("text/csv");
        response.setHeader(HttpHeaders.CONTENT_DISPOSITION, "attachment; filename=data.csv");

        return new CSVPrinter(response.getWriter(),
                CSVFormat.DEFAULT.withHeader(headers).withQuoteMode(QuoteMode.ALL));
    }

    /**
     * Writes a batch of source maps as CSV rows using the same field-extraction logic as
     * {@link #prepareToDownload}. Intended to be called repeatedly while paginating through
     * a large result set; pair with {@link #openCsvStream}.
     */
    public static void writeCsvBatch(CSVPrinter printer, DataColumn[] columns, List<?> data) throws IOException {
        if (data == null || data.isEmpty()) return;

        final DateTimeFormatter DATE_FORMATTER = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss z")
                .withLocale(Locale.getDefault()).withZone(TimezoneUtil.getAppTimezone());

        for (Object d : data) {
            DocumentContext ctx = JsonPath.parse(d);
            String[] cells = new String[columns.length];
            for (int i = 0; i < columns.length; i++) {
                String fieldName = columns[i].getField();
                String fieldType = columns[i].getType();
                cells[i] = null;

                Object value;
                try {
                    value = ctx.read("$." + fieldName);
                } catch (PathNotFoundException e) {
                    continue;
                }

                if (value == null)
                    continue;

                if (value instanceof String) {
                    cells[i] = "date".equals(fieldType) ? DATE_FORMATTER.format(Instant.parse(String.valueOf(value))) :
                            String.valueOf(value).replace("\n", " ").replace("\t", " ");
                } else if (value instanceof List) {
                    cells[i] = ((List<?>) value).stream().map(String::valueOf).collect(Collectors.joining(","));
                } else if (value instanceof Number) {
                    cells[i] = String.valueOf(value);
                } else if (value instanceof Map) {
                    try {
                        cells[i] = OBJECT_MAPPER.writeValueAsString(value);
                    } catch (Exception ex) {
                        cells[i] = value.toString();
                    }
                }
                cells[i] = sanitizeCsvCell(cells[i]);
            }
            printer.printRecord((Object[]) cells);
        }
        printer.flush();
    }

    /**
     * Neutralizes CSV-injection payloads by prefixing a single quote to any cell whose first
     * character is interpreted as a formula trigger by Excel/LibreOffice/Sheets.
     */
    private static String sanitizeCsvCell(String value) {
        if (value == null || value.isEmpty()) return value;
        char first = value.charAt(0);
        if (first == '=' || first == '+' || first == '-' || first == '@' || first == '\t' || first == '\r')
            return "'" + value;
        return value;
    }
}
