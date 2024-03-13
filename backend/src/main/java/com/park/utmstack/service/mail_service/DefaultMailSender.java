package com.park.utmstack.service.mail_service;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.shared_types.enums.EncryptionType;
import org.springframework.mail.javamail.JavaMailSender;
import org.springframework.mail.javamail.JavaMailSenderImpl;
import org.springframework.stereotype.Component;
import org.springframework.util.StringUtils;

import java.util.Properties;

@Component
public class DefaultMailSender extends BaseMailSender {

    public DefaultMailSender(){
        super(EncryptionType.NONE);
    }

    @Override
    public JavaMailSender getJavaMailSender() {
        JavaMailSender mailSender = super.getJavaMailSender();
        addProperties(mailSender);
        return mailSender;
    }

    @Override
    public JavaMailSender getJavaMailSender(String host, String username, String password, String protocol, String port){
        JavaMailSender mailSender = super.getJavaMailSender(host, username, password, protocol, StringUtils.hasText(port) ? port : Constants.PROP_EMAIL_PORT_NONE_VALUE.toString());
        addProperties(mailSender);
        return mailSender;
    }

    private void addProperties(JavaMailSender mailSender){

        Properties props = ((JavaMailSenderImpl) mailSender).getJavaMailProperties();

        props.clear();
        props.put("mail.smtp.auth", "false");
        props.put("mail.smtp.ssl.enable", "false");
        props.put("mail.smtp.ssl.trust", ((JavaMailSenderImpl) mailSender).getHost());

    }

}
