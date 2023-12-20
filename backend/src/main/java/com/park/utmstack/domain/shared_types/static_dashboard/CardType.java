package com.park.utmstack.domain.shared_types.static_dashboard;

public class CardType {
    private String serie;
    private Integer value;

    public CardType() {
    }

    public CardType(String serie, Integer value) {
        this.serie = serie;
        this.value = value;
    }

    public String getSerie() {
        return serie;
    }

    public void setSerie(String serie) {
        this.serie = serie;
    }

    public Integer getValue() {
        return value;
    }

    public void setValue(Integer value) {
        this.value = value;
    }
}
