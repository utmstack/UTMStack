package com.park.utmstack.domain.ip_info;

import org.springframework.util.StringUtils;

public class GeoIp {
    private String network;
    private String latitude;
    private String longitude;
    private String localeCode;
    private String continentCode;
    private String continentName;
    private String countryIsoCode;
    private String countryName;
    private String subdivision1IsoCode;
    private String subdivision1IsoName;
    private String subdivision2IsoCode;
    private String subdivision2IsoName;
    private String cityName;
    private String metroCode;
    private String timeZone;

    public String getNetwork() {
        return network;
    }

    public void setNetwork(String network) {
        this.network = network;
    }

    public Double getLatitude() {
        return Double.valueOf(latitude);
    }

    public void setLatitude(String latitude) {
        this.latitude = latitude;
    }

    public Double getLongitude() {
        return Double.valueOf(longitude);
    }

    public void setLongitude(String longitude) {
        this.longitude = longitude;
    }

    public String getLocaleCode() {
        return localeCode;
    }

    public void setLocaleCode(String localeCode) {
        this.localeCode = localeCode;
    }

    public String getContinentCode() {
        return continentCode;
    }

    public void setContinentCode(String continentCode) {
        this.continentCode = continentCode;
    }

    public String getContinentName() {
        return continentName;
    }

    public void setContinentName(String continentName) {
        this.continentName = continentName;
    }

    public String getCountryIsoCode() {
        return countryIsoCode;
    }

    public void setCountryIsoCode(String countryIsoCode) {
        this.countryIsoCode = countryIsoCode;
    }

    public String getCountryName() {
        return countryName;
    }

    public void setCountryName(String countryName) {
        this.countryName = countryName;
    }

    public String getSubdivision1IsoCode() {
        return subdivision1IsoCode;
    }

    public void setSubdivision1IsoCode(String subdivision1IsoCode) {
        this.subdivision1IsoCode = subdivision1IsoCode;
    }

    public String getSubdivision1IsoName() {
        return subdivision1IsoName;
    }

    public void setSubdivision1IsoName(String subdivision1IsoName) {
        this.subdivision1IsoName = subdivision1IsoName;
    }

    public String getSubdivision2IsoCode() {
        return subdivision2IsoCode;
    }

    public void setSubdivision2IsoCode(String subdivision2IsoCode) {
        this.subdivision2IsoCode = subdivision2IsoCode;
    }

    public String getSubdivision2IsoName() {
        return subdivision2IsoName;
    }

    public void setSubdivision2IsoName(String subdivision2IsoName) {
        this.subdivision2IsoName = subdivision2IsoName;
    }

    public String getCityName() {
        return cityName;
    }

    public void setCityName(String cityName) {
        this.cityName = cityName;
    }

    public String getMetroCode() {
        return metroCode;
    }

    public void setMetroCode(String metroCode) {
        this.metroCode = metroCode;
    }

    public String getTimeZone() {
        return timeZone;
    }

    public void setTimeZone(String timeZone) {
        this.timeZone = timeZone;
    }

    @Override
    public String toString() {
        return "Continent Name: " + (StringUtils.hasText(continentName) ? continentName : "-") + "\n" +
            "Continent Code: " + (StringUtils.hasText(continentCode) ? continentCode : "-") + "\n" +
            "Country Name: " + (StringUtils.hasText(countryName) ? countryName : "-") + "\n" +
            "Country ISO Code: " + (StringUtils.hasText(countryIsoCode) ? countryIsoCode : "-") + "\n" +
            "City Name: " + (StringUtils.hasText(cityName) ? cityName : "-") + "\n" +
            "Network: " + (StringUtils.hasText(network) ? network : "-") + "\n" +
            "Latitude: " + (StringUtils.hasText(latitude) ? latitude : "-") + "\n" +
            "Longitude: " + (StringUtils.hasText(longitude) ? longitude : "-") + "\n";
    }
}
