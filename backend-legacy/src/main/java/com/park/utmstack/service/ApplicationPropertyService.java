package com.park.utmstack.service;

import org.springframework.stereotype.Service;

@Service
public class ApplicationPropertyService {
//    private static final String CLASSNAME = "ApplicationPropertiesService";
//
//    private final ConfigurableEnvironment env;
//
//    public ApplicationPropertyService(ConfigurableEnvironment env) {
//        this.env = env;
//    }
//
//    public void updatePropertySource(String key, Object value) {
//        final String ctx = CLASSNAME + ".updatePropertySource";
//        try {
//            MutablePropertySources mutablePropertySources = env.getPropertySources();
//            MapPropertySource propertySources =
//                (MapPropertySource) mutablePropertySources.get(Constants.UTMSTACK_PROPERTY_SOURCE_NAME);
//            Objects.requireNonNull(propertySources).getSource().put(key, value);
//            mutablePropertySources.addFirst(propertySources);
//        } catch (Exception e) {
//            throw new RuntimeException(ctx + ": " + e.getMessage());
//        }
//    }
//
//    public String getProperty(String key) {
//        final String ctx = CLASSNAME + ".getProperty";
//        try {
//            MutablePropertySources propertySources = env.getPropertySources();
//            MapPropertySource map = (MapPropertySource) propertySources.get(Constants.UTMSTACK_PROPERTY_SOURCE_NAME);
//            Map<String, Object> source = Objects.requireNonNull(map).getSource();
//            return !Objects.isNull(source.get(key)) ? String.valueOf(source.get(key)) : "";
//        } catch (Exception e) {
//            throw new RuntimeException(ctx + ": " + e.getMessage());
//        }
//    }
}
