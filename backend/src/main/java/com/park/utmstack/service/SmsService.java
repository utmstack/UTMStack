package com.park.utmstack.service;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.twilio.Twilio;
import com.twilio.rest.api.v2010.account.Message;
import com.twilio.type.PhoneNumber;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Service;

import java.util.stream.Stream;

/**
 * Service layer to communicate with twilio api
 *
 * @author Leonardo M. Lopez
 */
@Service
public class SmsService {
    private static final String CLASS_NAME = "SmsService";
    private final Logger log = LoggerFactory.getLogger(SmsService.class);

    private boolean twilioConfigured;
    private final ApplicationEventService applicationEventService;

    public SmsService(ApplicationEventService applicationEventService) {
        this.applicationEventService = applicationEventService;
    }

    private void configureTwilio() {
        if (!twilioConfigured)
            Twilio.init(Constants.CFG.get(Constants.PROP_TWILIO_ACCOUNT_SID),
                Constants.CFG.get(Constants.PROP_TWILIO_AUTH_TOKEN));
        twilioConfigured = true;
    }

    /**
     * Send a sms message to specific phone numbers
     *
     * @param body:    Body of the message
     * @param numbers: The numbers to send the message
     */
    @Async
    public void sendTextMessage(String body, String... numbers) {
        final String ctx = CLASS_NAME + ".sendTextMessage";

        configureTwilio();
        PhoneNumber from = new PhoneNumber(Constants.CFG.get(Constants.PROP_TWILIO_NUMBER));
        Stream.of(numbers).forEach(number -> {
            Message message = Message.creator(from, new PhoneNumber(number), body).create();
            try {
                validateMessageResponse(message);
            } catch (Exception e) {
                String msg = String.format(
                    "%1$s: Fail to send sms with identifier: %2$s to number: %3$s. The status of the message is: %4$s with error message: %5$s",
                    ctx, message.getSid(), message.getTo(), message.getStatus().toString(), message.getErrorMessage());
                log.error(msg);
                applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            }
        });
    }

    /**
     * Validate a Twilio sms sent response
     *
     * @param message: A Twilio message object
     * @throws Exception In case of any error
     */
    private void validateMessageResponse(Message message) throws Exception {
        Message.Status status = message.getStatus();
        if (status.equals(Message.Status.FAILED) || status.equals(Message.Status.UNDELIVERED))
            throw new Exception(String.format(
                "The sending of the sms with identifier: %1$s to number: %2$s failed. The status of the message is: %3$s with error message: %4$s",
                message.getSid(), message.getTo(), message.getStatus(), message.getErrorMessage()));
    }
}
