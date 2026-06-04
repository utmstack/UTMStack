package com.park.utmstack.domain.index_policy;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.google.gson.annotations.SerializedName;

import java.util.Objects;

/**
 * <p>
 * Transitions define the conditions that need to be met for a state to change.
 * After all actions in the current state are completed, the policy starts checking the conditions for transitions.
 * </p>
 *
 * <p>
 * Transitions are evaluated in the order in which they are defined.
 * For example, if the conditions for the first transition are met,
 * then this transition takes place and the rest of the transitions are dismissed.
 * </p>
 *
 * <p>
 * If you don’t specify any conditions in a transition and leave it empty,
 * then it’s assumed to be the equivalent of always true.
 * This means that the policy transitions the index to this state the moment it checks.
 * </p>
 */
@JsonInclude(JsonInclude.Include.NON_NULL)
public class Transition {

    /**
     * The name of the state to transition to if the conditions are met
     * Required: Yes
     */
    @SerializedName(value = "state_name")
    private String stateName;

    /**
     * List the conditions for the transition
     * Required: Yes
     *
     * @see TransitionCondition
     */
    private TransitionCondition conditions;

    public Transition() {
    }

    public Transition(String stateName, TransitionCondition conditions) {
        this.stateName = stateName;
        this.conditions = conditions;
    }

    public Transition(String stateName) {
        this.stateName = stateName;
    }

    public String getStateName() {
        return stateName;
    }

    public void setStateName(String stateName) {
        this.stateName = stateName;
    }

    public TransitionCondition getConditions() {
        return conditions;
    }

    public void setConditions(TransitionCondition conditions) {
        this.conditions = conditions;
    }

    public static Builder builder() {
        return new Builder();
    }

    public static class Builder {
        private String stateName;
        private TransitionCondition conditions;

        public Builder stateName(String stateName) {
            this.stateName = stateName;
            return this;
        }

        public Builder ifMinIndexAgeIs(String time) {
            if (Objects.isNull(conditions))
                conditions = new TransitionCondition();
            this.conditions.setMinIndexAge(time);
            return this;
        }

        public Transition build() {
            return new Transition(stateName, conditions);
        }
    }
}
