package com.park.utmstack.util;


import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import com.park.utmstack.util.exceptions.UtmSerializationException;
import org.springframework.core.serializer.support.SerializationFailedException;

import java.io.*;
import java.util.List;

public class UtilSerializer {

    private static final String CLASS_NAME = "UtilSerializer";


    /**
     * Serialize an object
     *
     * @param obj : Object that will be serialized
     * @return A byte array with information of the serialized object
     * @throws UtmSerializationException
     * @author Leonardo M. Lopez
     */
    public static byte[] serialize(Object obj) throws UtmSerializationException {
        String ctx = CLASS_NAME + ".serialize";
        try {
            if (!(obj instanceof Serializable))
                throw new SerializationFailedException(
                    "This object can not being serialized because it is not an instance of Serializable class");
            ByteArrayOutputStream out = new ByteArrayOutputStream();
            ObjectOutputStream oos = new ObjectOutputStream(out);
            oos.writeObject(obj);
            return out.toByteArray();
        } catch (Exception e) {
            throw new UtmSerializationException(ctx + ": Fail to serialize, the message is: " + e.getMessage());
        }
    }

    /**
     * Deserialize a byte array with the information of an object that has being serialized previously
     *
     * @param array : Byte array that will be deserialized
     * @return An object deserialized from object
     * @throws UtmSerializationException
     * @throws ClassNotFoundException
     * @author Leonardo M. Lopez
     */
    public static Object deserialize(byte[] array) throws UtmSerializationException {
        String ctx = CLASS_NAME + ".deserialize";
        try {
            ObjectInputStream in = new ObjectInputStream(new ByteArrayInputStream(array));
            return in.readObject();
        } catch (Exception e) {
            throw new UtmSerializationException(ctx + ": Fail to deserialize, the message is: " + e.getMessage());
        }
    }

    public static <T> String jsonSerialize(T obj) throws UtmSerializationException {
        String ctx = CLASS_NAME + ".jsonSerialize";
        try {
            ObjectMapper om = new ObjectMapper();
            om.registerModule(new JavaTimeModule());
            return om.writeValueAsString(obj);
        } catch (Exception e) {
            throw new UtmSerializationException(ctx + ": Fail to serialize, the message is: " + e.getMessage());
        }
    }

    public static <T> T jsonDeserialize(Class<T> type, String json) throws UtmSerializationException {
        String ctx = CLASS_NAME + ".jsonDeserialize";
        try {
            ObjectMapper om = new ObjectMapper();
            om.registerModule(new JavaTimeModule());
            return om.readValue(json, type);
        } catch (Exception e) {
            throw new UtmSerializationException(ctx + ": Fail to deserialize, the message is: " + e.getMessage());
        }
    }

    public static <T> List<T> jsonDeserializeList(Class<T> type, String json) throws UtmSerializationException {
        String ctx = CLASS_NAME + ".jsonDeserialize";
        try {
            ObjectMapper om = new ObjectMapper();
            om.registerModule(new JavaTimeModule());
            return om.readValue(json,
                om.getTypeFactory().constructCollectionType(List.class, Class.forName(type.getName())));
        } catch (Exception e) {
            throw new UtmSerializationException(ctx + ": Fail to deserialize, the message is: " + e.getMessage());
        }
    }
}
