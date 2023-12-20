package com.park.utmstack.util;

import org.springframework.stereotype.Service;
import org.springframework.util.Assert;
import org.thymeleaf.context.Context;
import org.thymeleaf.exceptions.TemplateInputException;
import org.thymeleaf.spring5.SpringTemplateEngine;
import org.w3c.tidy.Tidy;
import org.xhtmlrenderer.pdf.ITextRenderer;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.UnsupportedEncodingException;
import java.nio.charset.StandardCharsets;
import java.util.Locale;
import java.util.Map;

@Service
public class PdfUtil {
    private static final String CLASS_NAME = "PdfUtil";

    private final SpringTemplateEngine templateEngine;

    public PdfUtil(SpringTemplateEngine templateEngine) {
        this.templateEngine = templateEngine;
    }

    /**
     * Using thymeleaf to create a html report, this method take the html processed and
     * create a pdf with it using flying-saucer-pdf lib
     *
     * @param htmlTemplate : Html-thymeleaf template uri
     * @param vars         : A map with variables needed by a html-thymeleaf template
     * @return A {@link ByteArrayOutputStream} object with pdf content
     * @throws Exception In case of any error
     */
    public ByteArrayOutputStream convertHtmlTemplateToPdf(String htmlTemplate, Map<String, Object> vars) throws Exception {
        final String ctx = CLASS_NAME + ".convertHtmlTemplateToPdf";
        Assert.hasText(htmlTemplate, ctx + ": Missing html template parameter");

        try {
            Locale locale = Locale.forLanguageTag("en");

            Context context = new Context(locale);
            context.setVariables(vars);

            String htmlContent;
            try {
                htmlContent = templateEngine.process(htmlTemplate, context);
                htmlContent = convertToXhtml(htmlContent);
            } catch (TemplateInputException e) {
                throw new Exception(e.getCause().getMessage());
            } catch (Exception e) {
                throw new Exception(e.getMessage());
            }

            //************BORRAR - just for testing
//            FileOutputStream fileOuputStream = new FileOutputStream("C:/temp/docs/report.html");
//            fileOuputStream.write(htmlContent.getBytes());
//            fileOuputStream.flush();
//            fileOuputStream.close();
            //*************************

            ByteArrayOutputStream bos = new ByteArrayOutputStream();

            ITextRenderer renderer = new ITextRenderer();

            renderer.setDocumentFromString(htmlContent);

            renderer.layout();
            renderer.createPDF(bos, false);
            renderer.finishPDF();
            bos.close();

            //************BORRAR - just for testing
//            FileOutputStream fileOuputStream1 = new FileOutputStream("C:/temp/docs/report.pdf");
//            fileOuputStream1.write(bos.toByteArray());
//            fileOuputStream1.flush();
//            fileOuputStream1.close();
            //*************************

            return bos;

        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    private String convertToXhtml(String html) throws UnsupportedEncodingException {
        Tidy tidy = new Tidy();
        tidy.setInputEncoding("UTF-8");
        tidy.setOutputEncoding("UTF-8");
        tidy.setXHTML(true);
        ByteArrayInputStream inputStream = new ByteArrayInputStream(html.getBytes(StandardCharsets.UTF_8));
        ByteArrayOutputStream outputStream = new ByteArrayOutputStream();
        tidy.parseDOM(inputStream, outputStream);
        return outputStream.toString(StandardCharsets.UTF_8);
    }
}
