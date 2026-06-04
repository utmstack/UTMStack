package com.park.utmstack.service.mail_sender;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.shared_types.enums.EncryptionType;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.mail.javamail.JavaMailSender;
import org.springframework.mail.javamail.JavaMailSenderImpl;
import org.springframework.stereotype.Component;
import org.springframework.util.StringUtils;

import java.util.Properties;

@Component
public class SslTlsMailSender extends BaseMailSender {

    @Autowired
    public SslTlsMailSender(JavaMailSenderImpl mailSender){
        super(EncryptionType.SSL_TLS, mailSender);
    }

    @Override
    public JavaMailSender getJavaMailSender() {
        JavaMailSender mailSender = super.getJavaMailSender();
        addProperties(mailSender);

        return mailSender;
    }

    @Override
    public JavaMailSender getJavaMailSender(String host, String username, String password, String protocol, Integer port){
        JavaMailSender mailSender = super.getJavaMailSender(host, username, password, protocol,  port != null && port != 0 ? port : Constants.PROP_EMAIL_PORT_SSL_VALUE);
        addProperties(mailSender);

        return mailSender;
    }

    private void addProperties(JavaMailSender mailSender){

        Properties props = ((JavaMailSenderImpl) mailSender).getJavaMailProperties();

        props.put("mail.smtp.auth", "true");
        props.put("mail.smtp.ssl.enable", "true");

    }

}
