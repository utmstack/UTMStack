package com.park.utmstack.domain.index_policy;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.google.gson.annotations.SerializedName;

import java.util.ArrayList;
import java.util.List;

/**
 * <p>Policies are JSON documents that define the following:</p>
 *
 * <li>
 * The states that an index can be in, including the default state for new indices.
 * For example, you might name your states “hot,” “warm,” “delete,” and so on.
 * For more information, see {@link State}.
 * </li>
 * <li>
 * Any actions that you want the plugin to take when an index enters a state, such as performing a rollover.
 * For more information, see {@link Action}.
 * </li>
 * <li>
 * The conditions that must be met for an index to move into a new state, known as transitions.
 * For example, if an index is more than eight weeks old, you might want to move it to the “delete” state.
 * For more information, see {@link Transition}.
 * </li>
 * <p>
 * In other words, a policy defines the states that an index can be in,
 * the actions to perform when in a state, and the conditions that must be met to transition between states.
 * </p>
 * <p>
 * You have complete flexibility in the way you can design your policies.
 * You can create any state, transition to any other state, and specify any number of actions in each state.
 * </p>
 */
public class Policy {
    /**
     * The name of the policy
     */
    @SerializedName(value = "policy_id")
    private String policyId;

    /**
     * A human-readable description of the policy
     */
    private String description;

    /**
     * The default starting state for each index that uses this policy
     */
    @SerializedName(value = "default_state")
    private String defaultState;

    /**
     * The states that you define in the policy
     *
     * @see State
     */
    private List<State> states = new ArrayList<>();

    @SerializedName(value = "ism_template")
    private List<IsmTemplate> ismTemplate;

    public Policy() {
    }

    public Policy(String defaultState, String description, List<State> states, List<IsmTemplate> ismTemplate) {
        this.description = description;
        this.defaultState = defaultState;
        this.states = states;
        this.ismTemplate = ismTemplate;
    }

    public String getPolicyId() {
        return policyId;
    }

    public void setPolicyId(String policyId) {
        this.policyId = policyId;
    }

    public String getDescription() {
        return description;
    }

    public void setDescription(String description) {
        this.description = description;
    }

    public String getDefaultState() {
        return defaultState;
    }

    public void setDefaultState(String defaultState) {
        this.defaultState = defaultState;
    }

    public List<State> getStates() {
        return states;
    }

    public void setStates(List<State> states) {
        this.states = states;
    }

    public List<IsmTemplate> getIsmTemplate() {
        return ismTemplate;
    }

    public void setIsmTemplate(List<IsmTemplate> ismTemplate) {
        this.ismTemplate = ismTemplate;
    }

    @JsonIgnore
    public static Builder builder() {
        return new Builder();
    }

    public static class Builder {
        private String defaultState;
        private String description;
        private final List<State> states = new ArrayList<>();
        private final List<IsmTemplate> ismTemplate = new ArrayList<>();

        public Builder withDefaultState(String defaultState) {
            this.defaultState = defaultState;
            return this;
        }

        public Builder withDescription(String description) {
            this.description = description;
            return this;
        }

        public Builder withState(State state) {
            this.states.add(state);
            return this;
        }

        public Builder withIsmTemplate(IsmTemplate ismTemplate) {
            this.ismTemplate.add(ismTemplate);
            return this;
        }

        public Policy build() {
            return new Policy(defaultState, description, states, ismTemplate);
        }
    }
}
