package com.park.utmstack.domain.index_policy;

public class RegisterRepositoryRequest {
    private final String type = "fs";
    private final RepositorySettings settings = new RepositorySettings();

    public RepositorySettings getSettings() {
        return settings;
    }
    public String getType() {
        return type;
    }

    public static class RepositorySettings {
        private final String location = "/usr/share/opensearch/backups";

        public String getLocation() {
            return location;
        }
    }
}
