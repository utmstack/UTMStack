package com.park.utmstack.service.mail_service;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.shared_types.enums.EncryptionType;
import org.springframework.mail.javamail.JavaMailSender;
import org.springframework.mail.javamail.JavaMailSenderImpl;

public abstract class BaseMailSender implements MailSenderStrategy {
    private final JavaMailSenderImpl mailSender;
    private final EncryptionType type;

    public BaseMailSender(EncryptionType type) {
        this.type = type;
        this.mailSender = new JavaMailSenderImpl();
    }

    private void configure(String host, String username, String password, String protocol, String port) {
        mailSender.setHost(host);
        mailSender.setPassword(password);
        mailSender.setUsername(username);
        mailSender.setProtocol(protocol);
        mailSender.setPort(Integer.parseInt(port));
    }

    @Override
    public JavaMailSender getJavaMailSender() {

        String host = Constants.CFG.get(Constants.PROP_MAIL_HOST);
        String username = Constants.CFG.get(Constants.PROP_MAIL_USERNAME);
        String password = Constants.CFG.get(Constants.PROP_MAIL_PASSWORD);
        String protocol = Constants.PROP_EMAIL_PROTOCOL_VALUE;
        String port = Constants.CFG.get(Constants.PROP_MAIL_PORT);

        configure(host, username, password, protocol, port);

        return mailSender;
    }


    @Override
    public JavaMailSender getJavaMailSender(String host, String username, String password, String protocol, String port) {

        configure(host, username, password, protocol, port);
        return mailSender;
    }

    @Override
    public String getEncryptionType(){
        return this.type.getType();
    }

}

