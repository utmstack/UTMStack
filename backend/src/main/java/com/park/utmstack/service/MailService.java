package com.park.utmstack.service;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.User;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.incident.UtmIncident;
import com.park.utmstack.domain.mail_sender.MailConfig;
import com.park.utmstack.domain.shared_types.AlertType;
import com.park.utmstack.domain.shared_types.LogType;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.mail_sender.BaseMailSender;
import com.park.utmstack.service.mail_sender.MailSenderStrategy;
import com.utmstack.opensearch_connector.types.ElasticCluster;
import org.apache.commons.csv.CSVFormat;
import org.apache.commons.csv.CSVPrinter;
import org.jetbrains.annotations.NotNull;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.context.MessageSource;
import org.springframework.core.io.ByteArrayResource;
import org.springframework.mail.javamail.JavaMailSender;
import org.springframework.mail.javamail.JavaMailSenderImpl;
import org.springframework.mail.javamail.MimeMessageHelper;
import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Service;
import org.springframework.util.CollectionUtils;
import org.springframework.util.StringUtils;
import org.thymeleaf.context.Context;
import org.thymeleaf.extras.java8time.dialect.Java8TimeDialect;
import org.thymeleaf.spring5.SpringTemplateEngine;

import javax.activation.DataHandler;
import javax.mail.BodyPart;
import javax.mail.Message;
import javax.mail.MessagingException;
import javax.mail.Multipart;
import javax.mail.internet.InternetAddress;
import javax.mail.internet.MimeBodyPart;
import javax.mail.internet.MimeMessage;
import javax.mail.internet.MimeMultipart;
import javax.mail.util.ByteArrayDataSource;
import java.io.ByteArrayOutputStream;
import java.io.InputStream;
import java.nio.charset.StandardCharsets;
import java.util.*;
import java.util.stream.Collectors;
import java.util.zip.ZipEntry;
import java.util.zip.ZipOutputStream;

/**
 * Service for sending emails.
 * <p>
 * We use the @Async annotation to send emails asynchronously.
 */
@Service
public class MailService {

    private final Logger log = LoggerFactory.getLogger(MailService.class);
    private static final String CLASS_NAME = "MailService";

    private static final String USER = "user";

    private static final String BASE_URL = "baseUrl";

    private final MessageSource messageSource;
    private final SpringTemplateEngine templateEngine;
    private final ApplicationEventService eventService;

    private final List<BaseMailSender> mailSenders;

    public MailService(MessageSource messageSource,
                       SpringTemplateEngine templateEngine,
                       ApplicationEventService eventService,
                       List<BaseMailSender> mailSenders) {
        this.messageSource = messageSource;
        this.templateEngine = templateEngine;
        this.eventService = eventService;
        this.mailSenders = mailSenders;
        this.templateEngine.addDialect(new Java8TimeDialect());
    }

//    private JavaMailSender getJavaMailSender() {
//        JavaMailSenderImpl mailSender = new JavaMailSenderImpl();
//        mailSender.setHost(Constants.CFG.get(Constants.PROP_MAIL_HOST));
//        mailSender.setPort(Integer.parseInt(Constants.CFG.get(Constants.PROP_MAIL_PORT)));
//        mailSender.setPassword(Constants.CFG.get(Constants.PROP_MAIL_PASSWORD));
//        mailSender.setUsername(Constants.CFG.get(Constants.PROP_MAIL_USERNAME));
//        mailSender.setProtocol(Constants.CFG.get(Constants.PROP_MAIL_PROTOCOL));
//
//        Properties props = mailSender.getJavaMailProperties();
//        props.put("mail.smtp.auth", Constants.CFG.get(Constants.PROP_MAIL_SMTP_AUTH));
//        props.put("mail.smtp.starttls.enable", Constants.CFG.get(Constants.PROP_MAIL_SMTP_STARTTLS_ENABLE));
//        props.put("mail.smtp.ssl.trust", Constants.CFG.get(Constants.PROP_MAIL_SMTP_SSL_TRUST));
//
//        return mailSender;
//    }

    private MailSenderStrategy getSender(String type){
        return mailSenders.stream()
                .filter(s -> s.getEncryptionType().equals(type))
                .findFirst()
                .orElseThrow(() -> new RuntimeException("Class not found"));
    }

    private @NotNull JavaMailSender getJavaMailSender(String host, String username, String password, String protocol, int port, String authType) throws MessagingException {

        MailSenderStrategy mailSenderStrategy = getSender(authType);
        JavaMailSender mailSender = mailSenderStrategy.getJavaMailSender(host, username, password, protocol, port);
        ((JavaMailSenderImpl) mailSender).testConnection();
        return mailSender;
    }

    private @NotNull JavaMailSender getJavaMailSender() throws MessagingException {

        String authType = Constants.CFG.get(Constants.PROP_MAIL_SMTP_AUTH);

        MailSenderStrategy mailSenderStrategy = getSender(authType);
        JavaMailSender mailSender = mailSenderStrategy.getJavaMailSender();
        ((JavaMailSenderImpl) mailSender).testConnection();
        return mailSender;
    }

    /**
     * Send a test email to the passed address to check if the email configuration is ok
     *
     * @param to Address to send the testing email
     */
    public void sendCheckEmail(List<String> to) throws MessagingException {
        try {
            JavaMailSender javaMailSender = getJavaMailSender();
            javaMailSender.send(this.getMimeMessage(javaMailSender, to, Constants.CFG.get(Constants.PROP_MAIL_FROM)));
        } catch (MessagingException e) {
            String msg = String.format("Email could not be sent to user(s) %1$s: The mail configuration is wrong: %2$s", String.join(",", to), e.getMessage());
            throw new MessagingException(msg);
        } catch (Exception e) {
            String msg = String.format("Email could not be sent to user(s) %1$s: %2$s", String.join(",", to), e.getMessage());
            throw new RuntimeException(msg);
        }
    }

    public void sendCheckEmail(List<String> to, MailConfig config) throws MessagingException {
        try {
            JavaMailSender javaMailSender = getJavaMailSender(config.getHost(), config.getUsername(), config.getPassword(), Constants.PROP_EMAIL_PROTOCOL_VALUE, config.getPort(), config.getAuthType());
            javaMailSender.send(this.getMimeMessage(javaMailSender, to, config.getFrom()));
        } catch (MessagingException e) {
            String msg = String.format("Email could not be sent to user(s) %1$s: The mail configuration is wrong: %2$s", String.join(",", to), e.getMessage());
            throw new MessagingException(msg);
        } catch (Exception e) {
            String msg = String.format("Email could not be sent to user(s) %1$s: %2$s", String.join(",", to), e.getMessage());
            throw new RuntimeException(msg);
        }
    }

    @Async
    public void sendEmail(List<String> to, String subject, String content, boolean isMultipart, boolean isHtml) {
        log.debug("Send email[multipart '{}' and html '{}'] to '{}' with subject '{}' and content={}", isMultipart, isHtml,
                to, subject, content);
        final String ctx = CLASS_NAME + ".sendEmail";
        try {
            JavaMailSender javaMailSender = getJavaMailSender();
            // Prepare message using a Spring helper
            MimeMessage mimeMessage = javaMailSender.createMimeMessage();

            MimeMessageHelper message = new MimeMessageHelper(mimeMessage, isMultipart, StandardCharsets.UTF_8.name());
            message.setTo(to.toArray(new String[]{}));
            message.setFrom(Constants.CFG.get(Constants.PROP_MAIL_FROM));
            message.setSubject(subject);
            message.setText(content, isHtml);
            javaMailSender.send(mimeMessage);
        } catch (MessagingException e) {
            String msg = String.format("%1$s: Email could not be sent to user(s) %2$s: The mail configuration is wrong: %3$s", ctx, String.join(",", to), e.getMessage());
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
        } catch (Exception e) {
            String msg = String.format("Email could not be sent to user(s) %1$s: %2$s", String.join(",", to), e.getMessage());
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
        }
    }

    @Async
    public void sendEmailWithAttachment(String emailsTo, String subject, String content, InputStream attach) {

        try {
            JavaMailSender javaMailSender = getJavaMailSender();
            // Prepare message using a Spring helper
            MimeMessage mimeMessage = javaMailSender.createMimeMessage();

            Multipart mail = new MimeMultipart();

            BodyPart text = new MimeBodyPart();
            text.setText(content);
            mail.addBodyPart(text);

            if (attach != null) {
                BodyPart attachment = new MimeBodyPart();
                attachment.setDataHandler(new DataHandler(new ByteArrayDataSource(attach, "application/pdf")));
                mail.addBodyPart(attachment);
            }

            mimeMessage.setFrom(Constants.CFG.get(Constants.PROP_MAIL_FROM));
            mimeMessage.setRecipients(Message.RecipientType.TO, InternetAddress.parse(emailsTo));
            mimeMessage.setSubject(subject);
            javaMailSender.send(mimeMessage);
        } catch (MessagingException e) {
            String msg = String.format("Email could not be sent to user(s) %1$s: The mail configuration is wrong: %2$s", emailsTo, e.getMessage());
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
        } catch (Exception e) {
            String msg = String.format("Email could not be sent to user(s) %1$s: %2$s", String.join(",", emailsTo), e.getMessage());
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
        }
    }

    @Async
    public void sendEmailFromTemplate(User user, String templateName, String titleKey) {
        Locale locale = Locale.forLanguageTag(user.getLangKey());
        Context context = new Context(locale);
        context.setVariable(USER, user);
        context.setVariable(BASE_URL, Constants.CFG.get(Constants.PROP_MAIL_BASE_URL));
        String content = templateEngine.process(templateName, context);
        String subject = messageSource.getMessage(titleKey, null, locale);
        sendEmail(Collections.singletonList(user.getEmail()), subject, content, false, true);
    }

    @Async
    public void sendActivationEmail(User user) {
        log.debug("Sending activation email to '{}'", user.getEmail());
        sendEmailFromTemplate(user, "mail/activationEmail", "email.activation.title");
    }

    @Async
    public void sendTfaVerificationCode(User user, String code) {
        Locale locale = Locale.forLanguageTag(user.getLangKey());
        Context context = new Context(locale);
        context.setVariable(USER, user);
        context.setVariable("tfaCode", code);
        String content = templateEngine.process("mail/tfaCodeEmail", context);
        String subject = "Two Factor Authentication Verification Code";
        sendEmail(Collections.singletonList(user.getEmail()), subject, content, false, true);
    }

    @Async
    public void sendLowSpaceEmail(List<User> users, ElasticCluster cluster) {
        Locale locale = Locale.forLanguageTag("en");
        Context context = new Context(locale);
        context.setVariable("cluster", cluster);
        context.setVariable("server", System.getenv(Constants.ENV_SERVER_NAME));
        String content = templateEngine.process("mail/elasticClusterStatusEmail", context);
        String subject = "Low space warning";
        sendEmail(users.stream().map(User::getEmail).collect(Collectors.toList()), subject, content, false, true);
    }

    @Async
    public void sendCreationEmail(User user) {
        log.debug("Sending creation email to '{}'", user.getEmail());
        sendEmailFromTemplate(user, "mail/creationEmail", "email.activation.title");
    }

    @Async
    public void sendPasswordResetMail(User user) {
        log.debug("Sending password reset email to '{}'", user.getEmail());
        sendEmailFromTemplate(user, "mail/passwordResetEmail", "email.reset.title");
    }

    @Async
    public void sendAlertEmail(List<String> emailsTo, AlertType alert, List<LogType> relatedLogs) {
        final String ctx = CLASS_NAME + ".sendAlertEmail";
        try {
            JavaMailSender javaMailSender = getJavaMailSender();
            if (CollectionUtils.isEmpty(emailsTo) || Objects.isNull(alert))
                return;

            Context context = new Context(Locale.ENGLISH);
            context.setVariable("alert", alert);
            context.setVariable("relatedLogs", relatedLogs);
            context.setVariable("timestamp", alert.getTimestampFormatted());
            context.setVariable(BASE_URL, Constants.CFG.get(Constants.PROP_MAIL_BASE_URL));

            final MimeMessage mimeMessage = javaMailSender.createMimeMessage();
            final MimeMessageHelper message = new MimeMessageHelper(mimeMessage, true, "UTF-8");
            message.setSubject(String.format("%1$s[%2$s]: %3$s", alert.getIncident() ? "INFOSEC-" : "", alert.getId().substring(0, 5), alert.getName()));
            message.setFrom(Constants.CFG.get(Constants.PROP_MAIL_FROM));
            message.setTo(emailsTo.toArray(new String[0]));

            final String htmlContent = templateEngine.process("mail/alertEmail", context);
            message.setText(htmlContent, true);

            ByteArrayResource zip = buildAlertEmailAttachment(context, alert, relatedLogs);
            message.addAttachment(String.format("Alert_%1$s.zip", alert.getId()), zip);

            javaMailSender.send(mimeMessage);
        } catch (MessagingException e) {
            String msg = String.format("%1$s: Email could not be sent to user(s) %2$s: The mail configuration is wrong: %3$s", ctx, String.join(",", emailsTo), e.getMessage());
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
        } catch (Exception e) {
            String msg = String.format("%1$s: Email could not be sent to %2$s: %3$s", ctx, String.join(",", emailsTo), e.getMessage());
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
        }
    }

    @Async
    public void sendIncidentEmail(List<String> emailsTo, List<AlertType> alerts, UtmIncident incident) {
        final String ctx = CLASS_NAME + ".sendIncidentEmail";
        try {
            JavaMailSender javaMailSender = getJavaMailSender();
            if (CollectionUtils.isEmpty(emailsTo) || CollectionUtils.isEmpty(alerts) || Objects.isNull(incident))
                return;

            Context context = new Context(Locale.ENGLISH);
            context.setVariable("alerts", alerts);
            context.setVariable("incident", incident);
            context.setVariable(BASE_URL, Constants.CFG.get(Constants.PROP_MAIL_BASE_URL));
            final MimeMessage mimeMessage = javaMailSender.createMimeMessage();
            final MimeMessageHelper message = new MimeMessageHelper(mimeMessage, true, "UTF-8");
            message.setSubject(String.format("%1$s[%2$s]: %3$s", "INFOSEC- New Incident ", incident.getId().toString(), incident.getIncidentName()));
            message.setFrom(Constants.CFG.get(Constants.PROP_MAIL_FROM));
            message.setTo(emailsTo.toArray(new String[0]));

            final String htmlContent = templateEngine.process("mail/newIncidentEmail", context);
            message.setText(htmlContent, true);

            javaMailSender.send(mimeMessage);
        } catch (MessagingException e) {
            String msg = String.format("%1$s: Email could not be sent to user(s) %2$s: The mail configuration is wrong: %3$s", ctx, String.join(",", emailsTo), e.getMessage());
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
        } catch (Exception e) {
            String msg = String.format("%1$s: Email could not be sent to %2$s: %3$s", ctx, String.join(",", emailsTo), e.getMessage());
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
        }
    }

    /**
     * Build attachment for alert mail
     *
     * @param context : Context used by spring template engine
     * @param alert   : Alert information
     * @return ByteArrayResource object with attachment to alert mail
     * @throws Exception In case of any error
     */
    private ByteArrayResource buildAlertEmailAttachment(Context context, AlertType alert,
                                                        List<LogType> relatedLogs) throws Exception {
        final String ctx = CLASS_NAME + ".buildAlertEmailAttachment";
        try {
            ByteArrayOutputStream bout = new ByteArrayOutputStream();
            ZipOutputStream zipOut = new ZipOutputStream(bout);
            zipOut.putNextEntry(new ZipEntry(String.format("%1$s.html", alert.getId())));
            zipOut.write(templateEngine.process("mail/alertEmailAttachment", context).getBytes(StandardCharsets.UTF_8));
            zipOut.closeEntry();

            if (!relatedLogs.isEmpty()) buildRelatedEventCsvAttachment(relatedLogs, zipOut);

            zipOut.close();
            return new ByteArrayResource(bout.toByteArray());
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    private void buildRelatedEventCsvAttachment(List<LogType> relatedLogs, ZipOutputStream zipOut) {
        final String ctx = CLASS_NAME + ".buildRelatedEventCsvAttachment";
        Map<String, List<LogType>> evtTypes = new HashMap<>();

        // Separating event types
        relatedLogs.forEach(doc -> {
            String logxType = doc.getDataType();

            evtTypes.computeIfAbsent(logxType, k -> new ArrayList<>());
            evtTypes.computeIfPresent(logxType, (k, v) -> {
                v.add(doc);
                return v;
            });
        });

        evtTypes.forEach((k, v) -> {
            // Extracting headers
            Set<String> set = new LinkedHashSet<>();
            v.forEach(value -> set.addAll(value.getLogxFlatted().keySet()));
            List<String> headers = new ArrayList<>(set);

            StringBuilder sb = new StringBuilder();
            try {
                CSVPrinter csvPrinter = new CSVPrinter(sb, CSVFormat.DEFAULT.withHeader(headers.toArray(new String[0])));
                v.forEach(value -> {
                    String[] cells = new String[headers.size()];

                    for (int i = 0; i < headers.size(); i++)
                        cells[i] = value.getLogxFlatted().computeIfPresent(headers.get(i), (kk, vv) -> vv);

                    try {
                        csvPrinter.printRecords((Object) cells);
                    } catch (Exception e) {
                        throw new RuntimeException(e);
                    }
                });
                zipOut.putNextEntry(new ZipEntry(String.format("%1$s.csv", k)));
                zipOut.write(sb.toString().getBytes(StandardCharsets.UTF_8));
                zipOut.closeEntry();
            } catch (Exception e) {
                throw new RuntimeException(ctx + ": " + e.getMessage());
            }
        });
    }

    @Async
    public void sendComplianceReportEmail(String emailTo, String subject, String content, String filename, byte[] attachment) {
        final String ctx = CLASS_NAME + ".sendComplianceReportEmail";
        try {
            JavaMailSender javaMailSender = getJavaMailSender();

            Context context = new Context(Locale.ENGLISH);
            context.setVariable(BASE_URL, Constants.CFG.get(Constants.PROP_MAIL_BASE_URL));
            context.setVariable("subject", subject);
            context.setVariable("content", content);

            final MimeMessage mimeMessage = javaMailSender.createMimeMessage();
            final MimeMessageHelper message = new MimeMessageHelper(mimeMessage, true, "UTF-8");
            message.setSubject(subject);
            message.setFrom(Constants.CFG.get(Constants.PROP_MAIL_FROM));
            message.setTo(emailTo);
            final String htmlContent = templateEngine.process("mail/complianceScheduleEmail", context);
            message.setText(htmlContent, true);

            ByteArrayResource complianceRep = new ByteArrayResource(attachment);
            message.addAttachment(filename, complianceRep);

            javaMailSender.send(mimeMessage);
        } catch (MessagingException e) {
            String msg = String.format("%1$s: Email could not be sent to user(s) %2$s: The mail configuration is wrong: %3$s", ctx, emailTo, e.getMessage());
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
        } catch (Exception e) {
            String msg = String.format("%1$s: Email could not be sent to %2$s: %3$s", ctx, emailTo, e.getMessage());
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
        }
    }

    private MimeMessage getMimeMessage(JavaMailSender javaMailSender, List<String> to, String from) throws MessagingException {
        MimeMessage mimeMessage = javaMailSender.createMimeMessage();
        MimeMessageHelper message = new MimeMessageHelper(mimeMessage, false, StandardCharsets.UTF_8.name());
        message.setTo(to.toArray(new String[]{}));
        message.setFrom(from);
        message.setSubject("Testing mail configuration");
        message.setText("Your email configuration is OK !!!");

        return mimeMessage;
    }
}
