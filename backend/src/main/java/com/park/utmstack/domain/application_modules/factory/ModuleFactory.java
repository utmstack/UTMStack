package com.park.utmstack.domain.application_modules.factory;

import com.park.utmstack.domain.application_modules.enums.ModuleName;
import com.park.utmstack.domain.application_modules.factory.impl.*;
import org.springframework.stereotype.Component;

@Component
public class ModuleFactory {
    private final ModuleFileIntegrity moduleFileIntegrity;
    private final ModuleO365 moduleO365;
    private final ModuleAzure moduleAzure;
    private final ModuleAwsCloudTrail moduleAwsCloudTrail;
    private final ModuleAwsIamUser moduleAwsIamUser;
    private final ModuleAwsTrafficMirror moduleAwsTrafficMirror;
    private final ModuleAwsBeanstalk moduleAwsBeanstalk;
    private final ModuleAwsFargate moduleAwsFargate;
    private final ModuleAwsLambda moduleAwsLambda;
    private final ModuleLinuxLogs moduleLinuxLogs;
    private final ModuleNetflow moduleNetflow;
    private final ModuleAwsPostgresql moduleAwsPostgresql;
    private final ModuleSophos moduleSophos;
    private final ModuleAwsSqlServer moduleAwsSqlServer;
    private final ModuleSyslog moduleSyslog;
    private final ModuleVMWare moduleVMWare;
    private final ModuleWindowsAgent moduleWindowsAgent;
    private final ModuleIIS moduleIIS;
    private final ModuleGcp moduleGcp;
    private final ModuleJson moduleJson;
    private final ModuleMacOsAgent moduleMacOsAgent;
    private final ModuleLinuxAgent moduleLinuxAgent;
    private final ModuleApache moduleApache;
    private final ModuleApache2 moduleApache2;
    private final ModuleAuditD moduleAuditD;
    private final ModuleElasticsearch moduleElasticsearch;
    private final ModuleHaProxy moduleHaProxy;
    private final ModuleKafka moduleKafka;
    private final ModuleKibana moduleKibana;
    private final ModuleLogstash moduleLogstash;
    private final ModuleMongoDb moduleMongoDb;
    private final ModuleMySql moduleMySql;
    private final ModuleNats moduleNats;
    private final ModuleNginx moduleNginx;
    private final ModuleOsQuery moduleOsQuery;
    private final ModulePostgreSql modulePostgreSql;
    private final ModuleRedis moduleRedis;
    private final ModuleTraefik moduleTraefik;
    private final ModuleCisco moduleCisco;
    private final ModuleMeraki moduleMeraki;
    private final ModuleEset moduleEset;
    private final ModuleKaspersky moduleKaspersky;
    private final ModuleSentinelOne moduleSentinelOne;
    private final ModuleFortiGate moduleFortiGate;
    private final ModuleSophosXg moduleSophosXg;
    private final ModuleFirePower moduleFirePower;
    private final ModuleUFW moduleUFW;
    private final ModuleMacOS moduleMacOS;
    private final ModuleMikrotik moduleMikrotik;
    private final ModulePaloAlto modulePaloAlto;
    private final ModuleCiscoSwitch moduleCiscoSwitch;
    private final ModuleSonicWall moduleSonicWall;
    private final ModuleDeceptiveBytes moduleDeceptiveBytes;
    private final ModuleGitHub moduleGitHub;
    private final ModuleBitdefender moduleBitdefender;
    private final ModuleSalesforce moduleSalesforce;
    private final ModuleIbmAs400 moduleIbmAs400;
    private final ModuleSocAi moduleSocAi;
    private final ModulePfsense modulePfsense;
    private final ModuleFortiWeb moduleFortiWeb;
    private final ModuleAix moduleAix;


    public ModuleFactory(ModuleFileIntegrity moduleFileIntegrity,
                         ModuleO365 moduleO365,
                         ModuleAzure moduleAzure,
                         ModuleAwsCloudTrail moduleAwsCloudTrail,
                         ModuleAwsIamUser moduleAwsIamUser,
                         ModuleAwsTrafficMirror moduleAwsTrafficMirror,
                         ModuleAwsBeanstalk moduleAwsBeanstalk,
                         ModuleAwsFargate moduleAwsFargate,
                         ModuleAwsLambda moduleAwsLambda,
                         ModuleLinuxLogs moduleLinuxLogs,
                         ModuleNetflow moduleNetflow,
                         ModuleAwsPostgresql moduleAwsPostgresql,
                         ModuleSophos moduleSophos,
                         ModuleAwsSqlServer moduleAwsSqlServer,
                         ModuleSyslog moduleSyslog,
                         ModuleVMWare moduleVMWare,
                         ModuleWindowsAgent moduleWindowsAgent,
                         ModuleIIS moduleIIS,
                         ModuleGcp moduleGcp,
                         ModuleJson moduleJson,
                         ModuleMacOsAgent moduleMacOsAgent,
                         ModuleLinuxAgent moduleLinuxAgent,
                         ModuleApache moduleApache,
                         ModuleApache2 moduleApache2,
                         ModuleAuditD moduleAuditD,
                         ModuleElasticsearch moduleElasticsearch,
                         ModuleHaProxy moduleHaProxy,
                         ModuleKafka moduleKafka,
                         ModuleKibana moduleKibana,
                         ModuleLogstash moduleLogstash,
                         ModuleMongoDb moduleMongoDb,
                         ModuleMySql moduleMySql,
                         ModuleNats moduleNats,
                         ModuleNginx moduleNginx,
                         ModuleOsQuery moduleOsQuery,
                         ModulePostgreSql modulePostgreSql,
                         ModuleRedis moduleRedis,
                         ModuleTraefik moduleTraefik,
                         ModuleCisco moduleCisco,
                         ModuleMeraki moduleMeraki,
                         ModuleEset moduleEset,
                         ModuleKaspersky moduleKaspersky,
                         ModuleSentinelOne moduleSentinelOne,
                         ModuleFortiGate moduleFortiGate,
                         ModuleSophosXg moduleSophosXg,
                         ModuleFirePower moduleFirePower,
                         ModuleUFW moduleUFW,
                         ModuleMacOS moduleMacOS,
                         ModuleMikrotik moduleMikrotik,
                         ModulePaloAlto modulePaloAlto,
                         ModuleCiscoSwitch moduleCiscoSwitch,
                         ModuleSonicWall moduleSonicWall,
                         ModuleDeceptiveBytes moduleDeceptiveBytes,
                         ModuleGitHub moduleGitHub,
                         ModuleBitdefender moduleBitdefender,
                         ModuleSalesforce moduleSalesforce,
                         ModuleIbmAs400 moduleIbmAs400,
                         ModuleSocAi moduleSocAi,
                         ModulePfsense modulePfsense,
                         ModuleFortiWeb moduleFortiWeb,
                         ModuleAix moduleAix) {
        this.moduleFileIntegrity = moduleFileIntegrity;
        this.moduleO365 = moduleO365;
        this.moduleAzure = moduleAzure;
        this.moduleAwsCloudTrail = moduleAwsCloudTrail;
        this.moduleAwsIamUser = moduleAwsIamUser;
        this.moduleAwsTrafficMirror = moduleAwsTrafficMirror;
        this.moduleAwsBeanstalk = moduleAwsBeanstalk;
        this.moduleAwsFargate = moduleAwsFargate;
        this.moduleAwsLambda = moduleAwsLambda;
        this.moduleLinuxLogs = moduleLinuxLogs;
        this.moduleNetflow = moduleNetflow;
        this.moduleAwsPostgresql = moduleAwsPostgresql;
        this.moduleSophos = moduleSophos;
        this.moduleAwsSqlServer = moduleAwsSqlServer;
        this.moduleSyslog = moduleSyslog;
        this.moduleVMWare = moduleVMWare;
        this.moduleWindowsAgent = moduleWindowsAgent;
        this.moduleIIS = moduleIIS;
        this.moduleGcp = moduleGcp;
        this.moduleJson = moduleJson;
        this.moduleMacOsAgent = moduleMacOsAgent;
        this.moduleLinuxAgent = moduleLinuxAgent;
        this.moduleApache = moduleApache;
        this.moduleApache2 = moduleApache2;
        this.moduleAuditD = moduleAuditD;
        this.moduleElasticsearch = moduleElasticsearch;
        this.moduleHaProxy = moduleHaProxy;
        this.moduleKafka = moduleKafka;
        this.moduleKibana = moduleKibana;
        this.moduleLogstash = moduleLogstash;
        this.moduleMongoDb = moduleMongoDb;
        this.moduleMySql = moduleMySql;
        this.moduleNats = moduleNats;
        this.moduleNginx = moduleNginx;
        this.moduleOsQuery = moduleOsQuery;
        this.modulePostgreSql = modulePostgreSql;
        this.moduleRedis = moduleRedis;
        this.moduleTraefik = moduleTraefik;
        this.moduleCisco = moduleCisco;
        this.moduleMeraki = moduleMeraki;
        this.moduleEset = moduleEset;
        this.moduleKaspersky = moduleKaspersky;
        this.moduleSentinelOne = moduleSentinelOne;
        this.moduleFortiGate = moduleFortiGate;
        this.moduleSophosXg = moduleSophosXg;
        this.moduleFirePower = moduleFirePower;
        this.moduleUFW = moduleUFW;
        this.moduleMacOS = moduleMacOS;
        this.moduleMikrotik = moduleMikrotik;
        this.modulePaloAlto = modulePaloAlto;
        this.moduleCiscoSwitch = moduleCiscoSwitch;
        this.moduleSonicWall = moduleSonicWall;
        this.moduleDeceptiveBytes = moduleDeceptiveBytes;
        this.moduleGitHub = moduleGitHub;
        this.moduleBitdefender = moduleBitdefender;
        this.moduleSalesforce = moduleSalesforce;
        this.moduleIbmAs400 = moduleIbmAs400;
        this.moduleSocAi = moduleSocAi;
        this.modulePfsense = modulePfsense;
        this.moduleFortiWeb = moduleFortiWeb;
        this.moduleAix = moduleAix;
    }

    public IModule getInstance(ModuleName nameShort) {
        if (nameShort.equals(ModuleName.FILE_INTEGRITY))
            return moduleFileIntegrity;
        if (nameShort.equals(ModuleName.O365))
            return moduleO365;
        if (nameShort.equals(ModuleName.AZURE))
            return moduleAzure;
        if (nameShort.equals(ModuleName.NETFLOW))
            return moduleNetflow;
        if (nameShort.equals(ModuleName.WINDOWS_AGENT))
            return moduleWindowsAgent;
        if (nameShort.equals(ModuleName.SYSLOG))
            return moduleSyslog;
        if (nameShort.equals(ModuleName.LINUX_LOGS))
            return moduleLinuxLogs;
        if (nameShort.equals(ModuleName.VMWARE))
            return moduleVMWare;
        if (nameShort.equals(ModuleName.AWS_TRAFFIC_MIRROR))
            return moduleAwsTrafficMirror;
        if (nameShort.equals(ModuleName.AWS_IAM_USER))
            return moduleAwsIamUser;
        if (nameShort.equals(ModuleName.AWS_CLOUDTRAIL))
            return moduleAwsCloudTrail;
        if (nameShort.equals(ModuleName.AWS_SQL_SERVER))
            return moduleAwsSqlServer;
        if (nameShort.equals(ModuleName.AWS_POSTGRESQL))
            return moduleAwsPostgresql;
        if (nameShort.equals(ModuleName.AWS_BEANSTALK))
            return moduleAwsBeanstalk;
        if (nameShort.equals(ModuleName.AWS_FARGATE))
            return moduleAwsFargate;
        if (nameShort.equals(ModuleName.AWS_LAMBDA))
            return moduleAwsLambda;
        if (nameShort.equals(ModuleName.SOPHOS))
            return moduleSophos;
        if (nameShort.equals(ModuleName.IIS))
            return moduleIIS;
        if (nameShort.equals(ModuleName.GCP))
            return moduleGcp;
        if (nameShort.equals(ModuleName.JSON))
            return moduleJson;
        if (nameShort.equals(ModuleName.MACOS_AGENT))
            return moduleMacOsAgent;
        if (nameShort.equals(ModuleName.LINUX_AGENT))
            return moduleLinuxAgent;
        if (nameShort.equals(ModuleName.APACHE))
            return moduleApache;
        if (nameShort.equals(ModuleName.APACHE2))
            return moduleApache2;
        if (nameShort.equals(ModuleName.AUDITD))
            return moduleAuditD;
        if (nameShort.equals(ModuleName.ELASTICSEARCH))
            return moduleElasticsearch;
        if (nameShort.equals(ModuleName.HAPROXY))
            return moduleHaProxy;
        if (nameShort.equals(ModuleName.KAFKA))
            return moduleKafka;
        if (nameShort.equals(ModuleName.KIBANA))
            return moduleKibana;
        if (nameShort.equals(ModuleName.LOGSTASH))
            return moduleLogstash;
        if (nameShort.equals(ModuleName.MONGODB))
            return moduleMongoDb;
        if (nameShort.equals(ModuleName.MYSQL))
            return moduleMySql;
        if (nameShort.equals(ModuleName.NATS))
            return moduleNats;
        if (nameShort.equals(ModuleName.NGINX))
            return moduleNginx;
        if (nameShort.equals(ModuleName.OSQUERY))
            return moduleOsQuery;
        if (nameShort.equals(ModuleName.POSTGRESQL))
            return modulePostgreSql;
        if (nameShort.equals(ModuleName.REDIS))
            return moduleRedis;
        if (nameShort.equals(ModuleName.TRAEFIK))
            return moduleTraefik;
        if (nameShort.equals(ModuleName.CISCO))
            return moduleCisco;
        if (nameShort.equals(ModuleName.MERAKI))
            return moduleMeraki;
        if (nameShort.equals(ModuleName.ESET))
            return moduleEset;
        if (nameShort.equals(ModuleName.KASPERSKY))
            return moduleKaspersky;
        if (nameShort.equals(ModuleName.SENTINEL_ONE))
            return moduleSentinelOne;
        if (nameShort.equals(ModuleName.FORTIGATE))
            return moduleFortiGate;
        if (nameShort.equals(ModuleName.SOPHOS_XG))
            return moduleSophosXg;
        if (nameShort.equals(ModuleName.FIRE_POWER))
            return moduleFirePower;
        if (nameShort.equals(ModuleName.UFW))
            return moduleUFW;
        if (nameShort.equals(ModuleName.MACOS))
            return moduleMacOS;
        if (nameShort.equals(ModuleName.MIKROTIK))
            return moduleMikrotik;
        if (nameShort.equals(ModuleName.PALO_ALTO))
            return modulePaloAlto;
        if (nameShort.equals(ModuleName.CISCO_SWITCH))
            return moduleCiscoSwitch;
        if (nameShort.equals(ModuleName.SONIC_WALL))
            return moduleSonicWall;
        if (nameShort.equals(ModuleName.DECEPTIVE_BYTES))
            return moduleDeceptiveBytes;
        if (nameShort.equals(ModuleName.GITHUB))
            return moduleGitHub;
        if (nameShort.equals(ModuleName.BITDEFENDER))
            return moduleBitdefender;
        if (nameShort.equals(ModuleName.SALESFORCE))
            return moduleSalesforce;
        if (nameShort.equals(ModuleName.IBM_AS_400))
            return moduleIbmAs400;
        if (nameShort.equals(ModuleName.SOC_AI))
            return moduleSocAi;
        if (nameShort.equals(ModuleName.PFSENSE))
            return modulePfsense;
        if (nameShort.equals(ModuleName.FORTIWEB))
            return moduleFortiWeb;
        if (nameShort.equals(ModuleName.AIX))
            return moduleAix;
        throw new RuntimeException("Unrecognized module " + nameShort.name());
    }
}
