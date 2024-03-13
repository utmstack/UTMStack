package com.park.utmstack.service.mail_service;

import com.park.utmstack.domain.shared_types.enums.EncryptionType;
import org.springframework.mail.javamail.JavaMailSender;
import org.springframework.mail.javamail.JavaMailSenderImpl;

public interface MailSenderStrategy {
    JavaMailSender getJavaMailSender();
    JavaMailSender getJavaMailSender(String host, String username, String password, String protocol, String port);

    String getEncryptionType();
}
