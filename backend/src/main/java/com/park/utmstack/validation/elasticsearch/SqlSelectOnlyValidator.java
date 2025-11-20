package com.park.utmstack.validation.elasticsearch;

import javax.validation.ConstraintValidator;
import javax.validation.ConstraintValidatorContext;
import java.util.regex.Pattern;
import java.util.Set;

public class SqlSelectOnlyValidator implements ConstraintValidator<SqlSelectOnly, String> {

    private static final Pattern START_PATTERN =
            Pattern.compile("(?is)^\\s*select\\b.*", Pattern.DOTALL);

    private static final Pattern FORBIDDEN_PATTERN =
            Pattern.compile("(?is)\\b(insert|update|delete|drop|alter|create|replace|truncate|" +
                    "merge|grant|revoke|exec|execute|commit|rollback|into)\\b");

    private static final Pattern COMMENT_PATTERN =
            Pattern.compile("(?s)(--.*?$|/\\*.*?\\*/)", Pattern.MULTILINE);

    private static final Set<String> ALLOWED_FUNCTIONS = Set.of(
            "COUNT", "AVG", "MIN", "MAX", "SUM"
    );

    private static final Set<String> KEYWORDS_TO_IGNORE = Set.of(
            "AS"
    );

    @Override
    public boolean isValid(String value, ConstraintValidatorContext context) {
        if (value == null || value.trim().isEmpty()) {
            return true;
        }

        String query = value.trim().replaceAll(";+$", "").trim();
        String upper = query.toUpperCase();

        if (!START_PATTERN.matcher(query).matches()) {
            return addConstraintViolation(context, "Query must start with SELECT");
        }

        if (FORBIDDEN_PATTERN.matcher(query).find()) {
            return addConstraintViolation(context, "Query contains forbidden SQL keywords");
        }

        if (COMMENT_PATTERN.matcher(query).find()) {
            return addConstraintViolation(context, "Query must not contain SQL comments (-- or /* */)");
        }

        if (query.contains(";")) {
            return addConstraintViolation(context, "Query must not contain internal semicolons");
        }

        if (!isBalancedQuotes(query)) {
            return addConstraintViolation(context, "Quotes are not balanced");
        }

        if (!isBalancedParentheses(query)) {
            return addConstraintViolation(context, "Parentheses are not balanced");
        }

        if (query.toUpperCase().matches("(?i).*FROM\\s+(GROUP|WHERE|ORDER|$).*")) {
            return addConstraintViolation(context, "FROM clause must contain a valid index or pattern");
        }

        if (hasMisplacedCommas(query)) {
            return addConstraintViolation(context, "Query contains misplaced commas");
        }

        for (String func : extractFunctions(upper)) {
            if (!ALLOWED_FUNCTIONS.contains(func)) {
                return addConstraintViolation(context, "Unsupported SQL function: " + func);
            }
        }

        if (upper.contains("HAVING") && !upper.contains("GROUP BY")) {
            return addConstraintViolation(context, "HAVING clause requires GROUP BY");
        }

        return true;
    }

    private boolean addConstraintViolation(ConstraintValidatorContext context, String msg) {
        context.disableDefaultConstraintViolation();
        context.buildConstraintViolationWithTemplate(msg).addConstraintViolation();
        return false;
    }

    private boolean isBalancedParentheses(String query) {
        int count = 0;
        for (char c : query.toCharArray()) {
            if (c == '(') count++;
            else if (c == ')') count--;
            if (count < 0) return false;
        }
        return count == 0;
    }

    private boolean isBalancedQuotes(String query) {
        int sq = 0;
        int dq = 0;
        boolean escaped = false;

        for (char c : query.toCharArray()) {
            if (escaped) { escaped = false; continue; }
            if (c == '\\') { escaped = true; continue; }
            if (c == '\'') sq++;
            else if (c == '"') dq++;
        }

        return (sq % 2 == 0) && (dq % 2 == 0);
    }

    private Set<String> extractFunctions(String upperQuery) {
        Pattern funcPattern = Pattern.compile("\\b([A-Z]+)\\s*\\(");
        var matcher = funcPattern.matcher(upperQuery);
        Set<String> funcs = new java.util.HashSet<>();
        while (matcher.find()) {
            String func = matcher.group(1);
            if (!ALLOWED_FUNCTIONS.contains(func) && !KEYWORDS_TO_IGNORE.contains(func)) {
                funcs.add(func);
            }
        }
        return funcs;
    }

    private boolean hasMisplacedCommas(String query) {
        String upperQuery = query.toUpperCase();

        if (upperQuery.startsWith("SELECT ,") || upperQuery.contains(",,"))
            return true;

        String selectPart = query.replaceAll("(?i)^SELECT\\s+", "")
                .replaceAll("(?i)\\s+FROM.*", "")
                .trim();
        String[] fields = selectPart.split(",");
        return fields.length < 2 && selectPart.contains(" ");
    }
}
