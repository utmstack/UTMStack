package com.park.utmstack.service.mail_sender;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.shared_types.enums.EncryptionType;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.mail.javamail.JavaMailSender;
import org.springframework.mail.javamail.JavaMailSenderImpl;
import org.springframework.stereotype.Component;

import java.util.Properties;

@Component
public class StartTlsMailSender extends BaseMailSender {

    @Autowired
    public StartTlsMailSender(JavaMailSenderImpl mailSender){

        super(EncryptionType.STARTTLS, mailSender);
    }

    @Override
    public JavaMailSender getJavaMailSender() {
        JavaMailSender mailSender = super.getJavaMailSender();
        addProperties(mailSender);
        return mailSender;
    }

    @Override
    public JavaMailSender getJavaMailSender(String host, String username, String password, String protocol, Integer port){
        JavaMailSender mailSender = super.getJavaMailSender(host, username, password, protocol, port != null && port != 0 ? port : Constants.PROP_EMAIL_PORT_TLS_VALUE);
        addProperties(mailSender);
        return mailSender;
    }

    private void addProperties(JavaMailSender mailSender){

        Properties props = ((JavaMailSenderImpl) mailSender).getJavaMailProperties();

        props.put("mail.smtp.auth", "true");
        props.put("mail.smtp.starttls.enable", "true");
        props.put("mail.smtp.starttls.required", "true");
    }

}
