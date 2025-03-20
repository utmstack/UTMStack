# UTMStack 10.7.0 Release Notes
## New Features and Improvements
- **Agent & Collector Dependencies**: agents and collectors now fetch their dependencies from the **agent-manager**, improving consistency and centralizing dependency management.

- **Agent Installation**: improved the installation messages for the agent to provide clearer instructions during the setup process.

- **Agent Service Cleanup**: removed unnecessary services to streamline the system and reduce overhead.

- **Error Recovery**: enhanced the agent's ability to recover from certain data streaming errors when interacting with the agent-manager, improving stability and fault tolerance.

- **Debug Mode for Agents**: Added a debug mode for agents, allowing better troubleshooting and logging for debugging purposes.

- **Certificate Verification Improvements**: Improved certificate verification in agents to enhance security and prevent connection issues.

- **Windows ARM64 Agent Support**: Added support for a Windows ARM64 agent, expanding compatibility to more architectures.