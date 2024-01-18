export class GettingStartedType {
    id: number;
    stepShort: GettingStartedStepEnum;
    stepOrder: number;
    completed: boolean;
}

export enum GettingStartedStepEnum {
    SET_ADMIN_USER = 'SET_ADMIN_USER',
    DASHBOARD_BUILDER = 'DASHBOARD_BUILDER',
    THREAT_MANAGEMENT = 'THREAT_MANAGEMENT',
    INTEGRATIONS = 'INTEGRATIONS',
    INVITE_USERS = 'INVITE_USERS',
    APPLICATION_SETTINGS = 'APPLICATION_SETTINGS'
}

export enum GettingStartedStepNameEnum {
    SET_ADMIN_USER = 'Setup your admin account',
    DASHBOARD_BUILDER = 'Dashboard builder and navigation',
    THREAT_MANAGEMENT = 'Alert management',
    INTEGRATIONS = 'Integrations',
    INVITE_USERS = 'Invite users',
    APPLICATION_SETTINGS = 'Application settings'
}

export enum GettingStartedStepIconPathEnum {
    DASHBOARD_BUILDER = 'assets/icons/system/WEBSITE_ANALYSIS.svg',
    THREAT_MANAGEMENT = 'assets/icons/system/PROMOTION.svg',
    INTEGRATIONS = 'assets/icons/system/SOLUTIONS.svg',
    INVITE_USERS = 'assets/icons/system/USER.svg',
    APPLICATION_SETTINGS = 'assets/icons/system/SETTINGS.svg'
}

export enum GettingStartedStepUrlEnum {
    DASHBOARD_BUILDER = '/creator/dashboard/builder',
    THREAT_MANAGEMENT = '/data/alert/view',
    INTEGRATIONS = '/integrations/explore',
    INVITE_USERS = '/management/user',
    APPLICATION_SETTINGS = '/app-management/settings/application-config'
}

export enum GettingStartedStepVideoPathEnum {
    DASHBOARD_BUILDER = 'assets/video/getting_started_dashboard_step.gif',
    THREAT_MANAGEMENT = 'assets/video/getting_started_alert_step.gif',
    INTEGRATIONS = 'assets/video/getting_started_integrations_step.gif',
    INVITE_USERS = 'assets/video/test.gif',
    APPLICATION_SETTINGS = 'assets/video/getting_started_app_settings.gif',
    APPLICATION_SAAS_SETTINGS = 'assets/video/getting_started_app_saas_settings.gif',
}
export enum GettingStartedStepDocURLEnum {
    DASHBOARD_BUILDER = 'UTMStackComponents/Dashboards/README.html',
    THREAT_MANAGEMENT = 'UTMStackComponents/Threath%20Managment/README.html',
    INTEGRATIONS = 'assets/video/getting_started_integrations_step.gif',
    INVITE_USERS = 'assets/video/test.gif',
    APPLICATION_SETTINGS = 'assets/video/getting_started_app_settings.gif'
}

