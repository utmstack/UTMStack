package com.park.utmstack.domain.index_policy;

import com.fasterxml.jackson.annotation.JsonIgnore;

import java.util.ArrayList;
import java.util.List;

/**
 * A state is the description of the status that the managed index is currently in.
 * A managed index can be in only one state at a time.
 * Each state has associated actions that are executed sequentially on entering a
 * state and transitions that are checked after all the actions have been completed.
 */
public class State {
    /**
     * The name of the state
     * Required: Yes
     */
    private String name;

    /**
     * The actions to execute after entering a state.
     * For more information, see {@link Action}
     * Required: Yes
     */
    private List<Action> actions;

    /**
     * The next states and the conditions required to transition to those states. If no transitions exist,
     * the policy assumes that it’s complete and can now stop managing the index.
     * For more information, see {@link Transition}.
     */
    private List<Transition> transitions;

    public State() {
    }

    public State(String name, List<Action> actions, List<Transition> transitions) {
        this.name = name;
        this.actions = actions;
        this.transitions = transitions;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public List<Action> getActions() {
        return actions;
    }

    public void setActions(List<Action> actions) {
        this.actions = actions;
    }

    public List<Transition> getTransitions() {
        return transitions;
    }

    public void setTransitions(List<Transition> transitions) {
        this.transitions = transitions;
    }

    @JsonIgnore
    public static Builder builder() {
        return new Builder();
    }

    public static class Builder {
        private String name;
        private final List<Action> actions = new ArrayList<>();
        private final List<Transition> transitions = new ArrayList<>();

        public Builder withName(String name) {
            this.name = name;
            return this;
        }

        public Builder executeAction(Action action) {
            this.actions.add(action);
            return this;
        }

        public Builder transitionTo(Transition transition) {
            this.transitions.add(transition);
            return this;
        }

        public State build() {
            return new State(name, actions, transitions);
        }
    }
}
