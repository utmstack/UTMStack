/*
 * Vendored from tech.jhipster:JHipster 7.3.1 (tech/jhipster/config/locale/AngularCookieLocaleResolver.java)
 * and adapted for Spring 6 (Spring Boot 3.1.5). JHipster 8 removed this class.
 *
 * Behavior unchanged: the Angular frontend stores the locale cookie value as a
 * URL-encoded, double-quoted string (e.g. %22en%22). Spring's CookieLocaleResolver
 * cannot parse the surrounding quotes, so this subclass strips them when reading
 * (parseLocaleValue) and re-adds them when writing (toLocaleValue) — the two
 * protected hooks Spring 6 exposes for exactly this purpose. (In Spring 5 these
 * were done via addCookie() + a custom parse method; that API no longer exists.)
 */
package tech.jhipster.config.locale;

import org.springframework.util.StringUtils;
import org.springframework.web.servlet.i18n.CookieLocaleResolver;

import java.util.Locale;

/**
 * Angular cookie saved the locale with a double quote (%22en%22). So the default
 * CookieLocaleResolver is not able to parse the locale. This class strips (and
 * re-adds) the surrounding %22 quotes.
 */
public class AngularCookieLocaleResolver extends CookieLocaleResolver {

    /** Constant <code>QUOTE="%22"</code> (URL-encoded double quote). */
    public static final String QUOTE = "%22";

    /**
     * Spring 7 removed {@code setCookieName(String)}; the cookie name is now set via the
     * {@code CookieLocaleResolver(String)} constructor.
     */
    public AngularCookieLocaleResolver(String cookieName) {
        super(cookieName);
    }

    /**
     * Read hook: Angular writes the cookie value quoted (e.g. %22en%22). Strip the
     * quotes before Spring parses the locale.
     */
    @Override
    protected Locale parseLocaleValue(String localeValue) {
        String v = localeValue;
        if (v != null) {
            v = StringUtils.replace(v, QUOTE, "").trim();
        }
        if (v == null || v.isEmpty() || "-".equals(v)) {
            return null;
        }
        return StringUtils.parseLocale(v);
    }

    /**
     * Write hook: re-add the Angular %22 quotes so the value the backend writes
     * (via LocaleChangeInterceptor / changeLocale) is what the frontend reader
     * expects.
     */
    @Override
    protected String toLocaleValue(Locale locale) {
        return QUOTE + super.toLocaleValue(locale) + QUOTE;
    }
}
