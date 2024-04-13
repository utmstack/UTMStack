package com.park.utmstack.service.mail_sender;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.shared_types.enums.EncryptionType;
import org.springframework.mail.javamail.JavaMailSender;
import org.springframework.mail.javamail.JavaMailSenderImpl;

import java.util.Properties;

public abstract class BaseMailSender implements MailSenderStrategy {
    private final JavaMailSenderImpl mailSender;
    private final EncryptionType type;

    public BaseMailSender(EncryptionType type, JavaMailSenderImpl javaMailSender) {
        this.type = type;
        this.mailSender = javaMailSender;
    }

    private void configure(String host, String username, String password, String protocol, Integer port) {
        mailSender.setHost(host);
        mailSender.setPassword(password);
        mailSender.setUsername(username);
        mailSender.setProtocol(protocol);
        mailSender.setPort(port);

        addProperties();
    }

    @Override
    public JavaMailSender getJavaMailSender() {

        String host = Constants.CFG.get(Constants.PROP_MAIL_HOST);
        String username = Constants.CFG.get(Constants.PROP_MAIL_USERNAME);
        String password = Constants.CFG.get(Constants.PROP_MAIL_PASSWORD);
        String protocol = Constants.PROP_EMAIL_PROTOCOL_VALUE;
        Integer port = Integer.parseInt(Constants.CFG.get(Constants.PROP_MAIL_PORT));

        configure(host, username, password, protocol, port);

        return mailSender;
    }


    @Override
    public JavaMailSender getJavaMailSender(String host, String username, String password, String protocol, Integer port) {
        configure(host, username, password, protocol, port);
        return mailSender;
    }

    public void addProperties(){
        Properties props = mailSender.getJavaMailProperties();
        props.clear();
        props.put("mail.smtp.timeout", "5000");
        props.put("mail.smtp.connectiontimeout", "5000");
        props.put("mail.smtp.ssl.trust", mailSender.getHost());
    }


    @Override
    public String getEncryptionType(){
        return this.type.getType();
    }

}

