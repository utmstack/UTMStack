import {AbstractControl, ValidationErrors, ValidatorFn} from '@angular/forms';

export class IpFormsValidators {

  static ipOrCidr(): ValidatorFn {
    return (control: AbstractControl): ValidationErrors | null => {
      if (!control.value) {
        return null;
      }

      const value = control.value.trim();

      if (value.includes('/')) {
        return IpFormsValidators.validateCIDR(value) ? null : { invalidCidr: true };
      }

      const isValidIPv4 = IpFormsValidators.validateIPv4(value);
      const isValidIPv6 = IpFormsValidators.validateIPv6(value);

      return (isValidIPv4 || isValidIPv6) ? null : { invalidIp: true };
    };
  }

  private static validateIPv4(ip: string): boolean {
    const ipv4Regex = /^(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})$/;
    const match = ip.match(ipv4Regex);

    if (!match) {
      return false;
    }

    for (let i = 1; i <= 4; i++) {
      const octet = parseInt(match[i], 10);
      if (octet < 0 || octet > 255) {
        return false;
      }
    }

    return true;
  }

  private static validateIPv6(ip: string): boolean {
    // tslint:disable-next-line:max-line-length
    const ipv6Regex = /^(([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))$/;

    return ipv6Regex.test(ip);
  }

  private static validateCIDR(cidr: string): boolean {
    const parts = cidr.split('/');

    if (parts.length !== 2) {
      return false;
    }

    const [ip, prefix] = parts;
    const prefixNum = parseInt(prefix, 10);

    const isIPv4 = ip.includes('.') && !ip.includes(':');
    const isIPv6 = ip.includes(':');

    if (isIPv4) {
      if (!IpFormsValidators.validateIPv4(ip)) {
        return false;
      }

      if (isNaN(prefixNum) || prefixNum < 0 || prefixNum > 32) {
        return false;
      }
    } else if (isIPv6) {

      if (!IpFormsValidators.validateIPv6(ip)) {
        return false;
      }

      if (isNaN(prefixNum) || prefixNum < 0 || prefixNum > 128) {
        return false;
      }
    } else {
      return false;
    }

    return true;
  }

}
