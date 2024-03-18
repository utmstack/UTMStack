import {Step} from '../shared/step';

export const VMWARESTEPS: Step[] = [
  {id: '1', name: 'Log in to your VMware vSphere Client.'},
  {id: '2', name: 'Select the host that manages your VMware inventory'},
  {id: '3', name: 'Click on the Configuration tab.'},
  {id: '4', name: 'From the Software panel, click on Advanced Settings'},
  {id: '5', name: 'In the navigation menu, click on Syslog.'},
  {id: '6', name: 'Configure values for the following parameters:', content: 'stepContent6'},
  ];
