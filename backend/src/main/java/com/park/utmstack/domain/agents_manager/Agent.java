package com.park.utmstack.domain.agents_manager;

import org.springframework.util.StringUtils;

public class Agent {
    private String id;
    private String status;
    private String name;
    private String registerIP;
    private String ip;
    private Os os;
    private String source;

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public String getStatus() {
        return status;
    }

    public void setStatus(String status) {
        this.status = status;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getRegisterIP() {
        return registerIP;
    }

    public void setRegisterIP(String registerIP) {
        this.registerIP = registerIP;
    }

    public String getIp() {
        return ip;
    }

    public void setIp(String ip) {
        this.ip = ip;
    }

    public Os getOs() {
        return os;
    }

    public void setOs(Os os) {
        this.os = os;
    }

    public Boolean getAlive() {
        return StringUtils.hasText(status) && status.equalsIgnoreCase("active");
    }

    public String getSource() {
        return source;
    }

    public void setSource(String source) {
        this.source = source;
    }

    public static class Os {
        private String arch;
        private String major;
        private String minor;
        private String name;
        private String platform;
        private String version;

        public String getArch() {
            return arch;
        }

        public void setArch(String arch) {
            this.arch = arch;
        }

        public String getMajor() {
            return major;
        }

        public void setMajor(String major) {
            this.major = major;
        }

        public String getMinor() {
            return minor;
        }

        public void setMinor(String minor) {
            this.minor = minor;
        }

        public String getName() {
            return name;
        }

        public void setName(String name) {
            this.name = name;
        }

        public String getPlatform() {
            return platform;
        }

        public void setPlatform(String platform) {
            this.platform = platform;
        }

        public String getVersion() {
            return version;
        }

        public void setVersion(String version) {
            this.version = version;
        }

        @Override
        public String toString() {
            return String.format("%1$s %2$s", name, version);
        }
    }
}
