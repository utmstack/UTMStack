
# UTMStack 10.5.0 Release Notes
This update focuses on integrating new functionalities, optimizing existing integrations, and addressing critical issues reported by our users. Our continuous monitoring and feedback process have guided these improvements, ensuring a better user experience and robust security measures.

## Summary of New Features and Improvements:
### New Integrations
- IBM AS400 Integration: This version introduces support for IBM AS400, allowing seamless log collection and integration into your existing UTMStack environment. The new integration leverages a new concept called "collector", which efficiently gathers logs and transmits them to the UTMStack instance for comprehensive analysis.
### Integration Enhancements
- AWS and Sophos Reworking: We have reworked the integrations for AWS and Sophos, transitioning them from Python to GoLang. This significant change enhances the performance and stability of these integrations, addressing previous bugs that hindered their connectivity. The transition to GoLang not only improves execution speed but also reduces resource consumption, providing a smoother and more reliable integration experience.
### Performance Improvements
- Agent Manager Optimization: We have identified and resolved an issue related to excessive memory usage by the Agent Manager. This fix optimizes resource utilization, ensuring more efficient operation and reducing the likelihood of performance bottlenecks.
### Security Enhancements:
- Enhanced Denial of Service Protection: We have implemented additional security measures to further strengthen our defenses against potential denial of service (DoS) attacks. These improvements ensure the continued resilience and robustness of our platform against various threat scenarios.
### General Improvements:
- Enhanced error handling and operational resilience across various modules.
- Improved user interface interactions for a more intuitive and user-friendly experience.
- Logging System Fixes: Corrected issues within the logging system to ensure more accurate and reliable log management and analysis.
  
This release represents our commitment to providing a robust, secure, and high-performance platform. We appreciate the feedback from our users, which has been instrumental in guiding these improvements. We continue to strive for excellence and welcome any further feedback to help us enhance UTMStack even more.
