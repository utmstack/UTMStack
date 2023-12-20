package com.park.utmstack.util;

import com.google.gson.JsonArray;
import com.google.gson.JsonElement;
import com.google.gson.JsonObject;
import com.google.gson.JsonParser;
import com.park.utmstack.domain.shared_types.DataColumn;
import com.park.utmstack.util.exceptions.UtmCsvException;
import io.jsonwebtoken.lang.Assert;
import org.apache.commons.csv.CSVFormat;
import org.apache.commons.csv.CSVPrinter;
import org.apache.commons.csv.QuoteMode;
import org.springframework.http.HttpHeaders;
import org.springframework.util.StringUtils;

import javax.servlet.http.HttpServletResponse;
import java.time.Instant;
import java.time.ZoneId;
import java.time.format.DateTimeFormatter;
import java.util.ArrayList;
import java.util.List;
import java.util.Locale;
import java.util.stream.Collectors;
import java.util.stream.Stream;
import java.util.stream.StreamSupport;

public class UtilCsv {
    private static final String CLASS_NAME = "UtilCsv";

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
        try {
            Assert.notEmpty(columns);
            Assert.notEmpty(data);

            JsonElement parse = JsonParser.parseString(UtilSerializer.jsonSerialize(data));
            JsonArray elements = parse.getAsJsonArray();
            List<String[]> rows = new ArrayList<>();

            elements.forEach(element -> {
                String[] cells = new String[columns.length];
                for (int i = 0; i < columns.length; i++) {
                    String field = columns[i].getField();
                    String type = columns[i].getType();
                    if (field.endsWith("keyword"))
                        field = field.replace(".keyword", "");
                    JsonElement value = getPath(field.split("\\."), (JsonObject) element);

                    cells[i] = null;
                    if (value != null && !value.isJsonNull()) {
                        if (value.isJsonObject() || value.isJsonPrimitive()) {
                            if (type.equals("date"))
                                cells[i] = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss").withLocale(Locale.getDefault()).withZone(
                                    ZoneId.systemDefault()).format(Instant.parse(value.getAsString()));
                            else
                                cells[i] = value.getAsString()
                                    .replace("\n", " ").replace("\t", " ");
                        } else if (value.isJsonArray()) {
                            cells[i] = StreamSupport.stream(value.getAsJsonArray().spliterator(), false)
                                .map(JsonElement::getAsString).collect(Collectors.joining(","));
                        }
                    }
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

    private static JsonElement getPath(String[] path, JsonObject jsonObject) {
        if (jsonObject == null)
            return null;
        JsonElement obj = jsonObject;
        for (String component : path) {
            if (obj == null)
                break;
            obj = ((JsonObject) obj).get(component);
        }
        return obj;
    }
}
