package com.park.utmstack.service.mail_sender;

import org.springframework.mail.javamail.JavaMailSender;

public interface MailSenderStrategy {
    JavaMailSender getJavaMailSender();
    JavaMailSender getJavaMailSender(String host, String username, String password, String protocol, Integer port);

    String getEncryptionType();
}
