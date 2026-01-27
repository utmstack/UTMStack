package com.park.utmstack.domain.compliance.enums;

public enum EvaluationRule {
    NO_HITS_ALLOWED, // no results
    MIN_HITS_REQUIRED, // at least N results
    THRESHOLD_MAX, // a maximum of N results.
    MATCH_FIELD_VALUE // a specific value in a field
}