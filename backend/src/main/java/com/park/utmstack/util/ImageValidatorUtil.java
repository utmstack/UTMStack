package com.park.utmstack.util;

import com.park.utmstack.util.enums.ImageComponents;
import org.apache.tika.Tika;
import org.apache.tika.mime.MimeType;
import org.apache.tika.mime.MimeTypeException;
import org.apache.tika.mime.MimeTypes;

import javax.imageio.ImageIO;
import java.awt.image.BufferedImage;
import java.io.ByteArrayInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.util.*;

public class ImageValidatorUtil {

    public static final List<String> ALLOWED_EXTENSIONS = Arrays.asList(".png", ".jpg", ".jpeg");
    public static final List<String> ALLOWED_MIME_TYPES = Arrays.asList("image/png", "image/jpeg", "image/jpg");
    private static final Tika tika = new Tika();

    /**
     * Method to extract base64 data part
     * */
    private static String extractBase64Data(String base64Image) {
        return base64Image.split(",")[1]; // Extract base64 data part
    }

    /**
     * Returns the extension for a given mimetype
     * */
    private static String getExtensionFromMime(String mimeType) throws MimeTypeException {
        MimeTypes allTypes = MimeTypes.getDefaultMimeTypes();
        MimeType type = allTypes.forName(mimeType);
        return type.getExtension();
    }

    /**
     * Method to check if the image data is a real image to avoid XSS attack
     * */
    public static boolean isRealImage(InputStream t) throws IOException {
        BufferedImage bufferedImage = ImageIO.read(new ByteArrayInputStream(t.readAllBytes()));
        return bufferedImage != null;
    }

    /**
     * Receives image data in base 64 and return map with: mimeType, fileExtension, inputStream
     * The corresponding keys used are in the enum ImageComponents
     * */
    public static Map<ImageComponents,Object> imageComponents(String base64Data) throws IOException, MimeTypeException {
        Map<ImageComponents,Object> components = new LinkedHashMap<>();
        byte[] imageData = Base64.getDecoder().decode(extractBase64Data(base64Data));

        try (ByteArrayInputStream inputStream = new ByteArrayInputStream(imageData)) {

            String mimeType = tika.detect(inputStream);
            String fileExtension = getExtensionFromMime(mimeType);

            components.put(ImageComponents.IMG_MIME_TYPE,mimeType);
            components.put(ImageComponents.IMG_FILE_EXTENSION,fileExtension);
            components.put(ImageComponents.IMG_DATA,inputStream);
        }
        return components;
    }

    /**
     * Method to check if extension is allowed
     * */
    public static boolean isExtensionAllowed(String extension) {
        return ALLOWED_EXTENSIONS.contains(extension.toLowerCase(Locale.ROOT));
    }

    /**
     * Method to check if mimetype is allowed
     * */
    public static boolean isMimeTypeAllowed(String mimeType) {
        return ALLOWED_MIME_TYPES.contains(mimeType.toLowerCase(Locale.ROOT));
    }
}

