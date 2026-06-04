export class SDDLParser {
  aceTypes: { [key: string]: string; } = {};
  aceFlags: { [key: string]: string; } = {};
  acePermissions: { [key: string]: string; } = {};
  aceTrustee: { [key: string]: string; } = {};

  constructor() {
    this.initialize();
  }

  initialize() {
    // #region Add this.ACE_Types
    this.aceTypes.A = 'Access Allowed';
    this.aceTypes.D = 'Access Denied';
    this.aceTypes.OA = 'Object Access Allowed';
    this.aceTypes.OD = 'Object Access Denied';
    this.aceTypes.AU = 'System Audit';
    this.aceTypes.AL = 'System Alarm';
    this.aceTypes.OU = 'Object System Audit';
    this.aceTypes.OL = 'Object System Alarm';
    // #endregion
    // #region Add this.ACE_Flags
    this.aceFlags.CI = 'Container Inherit';
    this.aceFlags.OI = 'Object Inherit';
    this.aceFlags.NP = 'No Propagate';
    this.aceFlags.IO = 'Inheritance Only';
    this.aceFlags.ID = 'Inherited';
    this.aceFlags.SA = 'Successful Access Audit';
    this.aceFlags.FA = 'Failed Access Audit';
    // #endregion
    // #region Add this.Permissions
    // #region Generic Access Rights
    this.acePermissions.GA = 'Generic All';
    this.acePermissions.GR = 'Generic Read';
    this.acePermissions.GW = 'Generic Write';
    this.acePermissions.GX = 'Generic Execute';
    // #endregion
    // #region Directory Access Rights
    this.acePermissions.RC = 'Read this.Permissions';
    this.acePermissions.SD = 'Delete';
    this.acePermissions.WD = 'Modify this.Permissions';
    this.acePermissions.WO = 'Modify Owner';
    this.acePermissions.RP = 'Read All Properties';
    this.acePermissions.WP = 'Write All Properties';
    this.acePermissions.CC = 'Create All Child Objects';
    this.acePermissions.DC = 'Delete All Child Objects';
    this.acePermissions.LC = 'List Contents';
    this.acePermissions.SW = 'All Validated Writes';
    this.acePermissions.LO = 'List Object';
    this.acePermissions.DT = 'Delete Subtree';
    this.acePermissions.CR = 'All Extended Rights';
    // #endregion
    // #region File Access Rights
    this.acePermissions.FA = 'File All Access';
    this.acePermissions.FR = 'File Generic Read';
    this.acePermissions.FW = 'File Generic Write';
    this.acePermissions.FX = 'File Generic Execute';
    // #endregion
    // #region Registry Key Access Rights
    this.acePermissions.KA = 'Key All Access';
    this.acePermissions.KR = 'Key Read';
    this.acePermissions.KW = 'Key Write';
    this.acePermissions.KX = 'Key Execute';
    // #endregion
    // #endregion
    // #region Add this.Trustee's
    this.aceTrustee.AO = 'Account Operators';
    this.aceTrustee.RU = 'Alias to allow previous Windows 2000';
    this.aceTrustee.AN = 'Anonymous Logon';
    this.aceTrustee.AU = 'Authenticated Users';
    this.aceTrustee.BA = 'Built-in Administrators';
    this.aceTrustee.BG = 'Built in Guests';
    this.aceTrustee.BO = 'Backup Operators';
    this.aceTrustee.BU = 'Built-in Users';
    this.aceTrustee.CA = 'Certificate Server Administrators';
    this.aceTrustee.CG = 'Creator Group';
    this.aceTrustee.CO = 'Creator Owner';
    this.aceTrustee.DA = 'Domain Administrators';
    this.aceTrustee.DC = 'Domain Computers';
    this.aceTrustee.DD = 'Domain Controllers';
    this.aceTrustee.DG = 'Domain Guests';
    this.aceTrustee.DU = 'Domain Users';
    this.aceTrustee.EA = 'Enterprise Administrators';
    this.aceTrustee.ED = 'Enterprise Domain Controllers';
    this.aceTrustee.WD = 'Everyone';
    this.aceTrustee.PA = 'Group Policy Administrators';
    this.aceTrustee.IU = 'Interactively logged-on user';
    this.aceTrustee.LA = 'Local Administrator';
    this.aceTrustee.LG = 'Local Guest';
    this.aceTrustee.LS = 'Local Service Account';
    this.aceTrustee.SY = 'Local System';
    this.aceTrustee.NU = 'Network Logon User';
    this.aceTrustee.NO = 'Network Configuration Operators';
    this.aceTrustee.NS = 'Network Service Account';
    this.aceTrustee.PO = 'Printer Operators';
    this.aceTrustee.PS = 'Self';
    this.aceTrustee.PU = 'Power Users';
    this.aceTrustee.RS = 'RAS Servers group';
    this.aceTrustee.RD = 'Terminal Server Users';
    this.aceTrustee.RE = 'Replicator';
    this.aceTrustee.RC = 'Restricted Code';
    this.aceTrustee.SA = 'Schema Administrators';
    this.aceTrustee.SO = 'Server Operators';
    this.aceTrustee.SU = 'Service Logon User';
    // #endregion
  }

  friendlyTrusteeName(trustee: string) {
    if (this.aceTrustee[trustee]) {
      return this.aceTrustee[trustee];
    } else {
      return '';
    }
  }

  doParse(subSDDL: string, separator: string, separator2: string) {
    let retval = '';
    const type = subSDDL.charAt(0);

    if (type === 'O') {
      const owner = subSDDL.substr(2);
      return 'Owner: ' + this.friendlyTrusteeName(owner) + separator;
    } else if (type === 'G') {
      const group = subSDDL.substr(2);
      return 'Group: ' + this.friendlyTrusteeName(group) + separator;
    } else if ((type === 'D') || (type === 'S')) {
      if (type === 'D') {
        retval += 'DACL' + separator;
      } else {
        retval += 'SACL' + separator;
      }

      const sections: string[] = subSDDL.split('(');
      for (let count = 1; count < sections.length; count++) {
        retval += '# ' + count.toString() + ' of ' + (sections.length - 1).toString() + separator;
        const parts: string[] = sections[count].replace(')', '').split(';');
        retval += '';
        if (this.aceFlags[parts[0]]) {
          retval += separator2 + 'Type: ' + this.aceTypes[parts[0]] + separator;
        }
        if (this.aceFlags[parts[1]]) {
          retval += separator2 + 'Inheritance: ' + this.aceFlags[parts[1]] + separator;
        }
        for (let count2 = 0; count2 < parts[2].length; count2 += 2) {
          const perm = parts[2].substring(count2, 2);
          if (this.acePermissions[perm]) {
            if (count2 === 0) {
              retval += separator2 + 'Permissions: ' + this.acePermissions[perm];
            } else {
              retval += '|' + this.acePermissions[perm];
            }
          }
        }
        retval += separator;
        retval += separator2 + 'Trustee: ' + this.friendlyTrusteeName(parts[5]) + separator;
      }
    }
    return retval;
  }

  parse(SDDL: string, separator: string = '<br>', separator2: string = '&nbsp;&nbsp;&nbsp;&nbsp;') {
    let retval = '';
    if (this.aceTypes === {}) {
      this.initialize();
    }
    let startindex = 0;
    let nextindex = 0;
    let first = 0;
    let section = '';
    while (true) {
      first = SDDL.indexOf(':', nextindex) - 1;
      startindex = nextindex;
      if (first < 0) {
        break;
      }
      if (first !== 0) {
        section = SDDL.substring(startindex - 2, first - startindex + 2);
        retval += this.doParse(section, separator, separator2);
      }
      nextindex = first + 2;
    }
    section = SDDL.substring(startindex - 2);
    retval += this.doParse(section, separator, separator2);
    return retval;
  }

}
