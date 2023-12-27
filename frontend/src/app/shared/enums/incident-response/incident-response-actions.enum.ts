export enum IncidentResponseActionsEnum {
  SHUTDOWN_SERVER = 1, // shutdown server
  DISABLE_USER = 2, // kick out and disable user
  BLOCK_IP = 3, // # block ip and disconnect any traffic from IP -- it is possible to specify ranges using CIDR notation here.
  ISOLATE_HOST = 4, // Isolate host (disconnect from network)
  RESTART_SERVER = 5, // restart server
  KILL_PROCESS = 6, // kill process
  UNINSTALL_PROGRAM = 7, // uninstall program
  RUN_CMD = 8, // run shell command
}
