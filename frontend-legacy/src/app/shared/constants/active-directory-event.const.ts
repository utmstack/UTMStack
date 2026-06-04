export const ACTIVE_DIRECTORY_EVENT: { eventId: number, message: string, type: string }[] = [
  {eventId: 4741, message: 'A computer account was created', type: 'COMPUTER'},
  {eventId: 4742, message: 'A computer account was changed', type: 'COMPUTER'},
  {eventId: 4743, message: 'A computer account was deleted', type: 'COMPUTER'},
  {eventId: 4627, message: 'Group membership information.', type: 'GROUP'},
  {eventId: 4727, message: 'A security-enabled global group was created', type: 'GROUP'},
  {eventId: 4728, message: 'A member was added to a security-enabled global group', type: 'GROUP'},
  {eventId: 4729, message: 'A member was removed from a security-enabled global group', type: 'GROUP'},
  {eventId: 4730, message: 'A security-enabled global group was deleted', type: 'GROUP'},
  {eventId: 4731, message: 'A security-enabled local group was created', type: 'GROUP'},
  {eventId: 4732, message: 'A member was added to a security-enabled local group', type: 'GROUP'},
  {eventId: 4733, message: 'A member was removed from a security-enabled local group', type: 'GROUP'},
  {eventId: 4734, message: 'A security-enabled local group was deleted', type: 'GROUP'},
  {eventId: 4735, message: 'A security-enabled local group was changed', type: 'GROUP'},
  {eventId: 4737, message: 'A security-enabled global group was changed', type: 'GROUP'},
  {eventId: 4754, message: 'A security-enabled universal group was created', type: 'GROUP'},
  {eventId: 4755, message: 'A security-enabled universal group was changed', type: 'GROUP'},
  {eventId: 4756, message: 'A member was added to a security-enabled universal group', type: 'GROUP'},
  {eventId: 4757, message: 'A member was removed from a security-enabled universal group', type: 'GROUP'},
  {eventId: 4758, message: 'A security-enabled universal group was deleted', type: 'GROUP'},
  {eventId: 4764, message: 'A groups type was changed', type: 'GROUP'},
  {eventId: 4660, message: 'An object was deleted', type: 'OBJECT'},
  {eventId: 4662, message: 'An operation was performed on an object', type: 'OBJECT'},
  {eventId: 4670, message: 'Permissions on an object were changed', type: 'OBJECT'},
  {eventId: 4719, message: 'System audit policy was changed', type: 'OBJECT'},
  {eventId: 4739, message: 'Domain Policy was changed', type: 'OBJECT'},
  {
    eventId: 4611,
    message: 'A trusted logon process has been registered with the Local Security Authority',
    type: 'USER'
  },
  {eventId: 4624, message: 'An account was successfully logged on', type: 'USER'},
  {eventId: 4625, message: 'An account failed to log on ', type: 'USER'},
  {eventId: 4626, message: 'User/Device claims information', type: 'USER'},
  {eventId: 4634, message: 'An account was logged off  ', type: 'USER'},
  {eventId: 4647, message: 'User initiated logoff', type: 'USER'},
  {eventId: 4648, message: 'A logon was attempted using explicit credentials', type: 'USER'},
  {eventId: 4672, message: 'Special privileges assigned to new logon', type: 'USER'},
  {eventId: 4704, message: 'A user right was assigned', type: 'USER'},
  {eventId: 4705, message: 'A user right was removed', type: 'USER'},
  {eventId: 4717, message: 'System security access was granted to an account', type: 'USER'},
  {eventId: 4718, message: 'System security access was removed from an account', type: 'USER'},
  {eventId: 4720, message: 'A user account was created', type: 'USER'},
  {eventId: 4722, message: 'A user account was enabled', type: 'USER'},
  {eventId: 4723, message: 'An attempt was made to change an account\'s password', type: 'USER'},
  {eventId: 4724, message: 'An attempt was made to reset an accounts password', type: 'USER'},
  {eventId: 4725, message: 'A user account was disabled', type: 'USER'},
  {eventId: 4726, message: 'A user account was deleted', type: 'USER'},
  {eventId: 4738, message: 'A user account was changed', type: 'USER'},
  {eventId: 4740, message: 'A user account was locked out', type: 'USER'},
  {eventId: 4767, message: 'A user account was unlocked', type: 'USER'},
  {eventId: 4781, message: 'The name of an account was changed', type: 'USER'},
  {eventId: 4782, message: 'The password hash an account was accessed', type: 'USER'}
];


export function searchEventType(eventId: number): { eventId: number, message: string, type: string } {
  return ACTIVE_DIRECTORY_EVENT.find(value => {
    return value.eventId === eventId;
  });
}

// 4741, 4742, 4743, 4627, 4727, 4728, 4729, 4730, 4731, 4732,
// 4733, 4734, 4735, 4737, 4754, 4755, 4756, 4757, 4758, 4764,
// 4660, 4670, 4719, 4739, 4626, 4704, 4705, 4717, 4718, 4720,
// 4722, 4723, 4724, 4725, 4726, 4738, 4740, 4767, 4781, 4782
