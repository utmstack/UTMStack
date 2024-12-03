# New Workflows

### Branches involved
***main***<br>
***feature/\*\*/\*\****<br>
***hotfix/\*\*/\*\****<br>

## Feature:
1. **Create Branch**: Start a new branch from main for the feature.
2. **Develop**: All developers related to the feature work on this branch. 
3. Push Changes: When they finish they push.
4. **Activate DEV Actions**: This push triggers DEV actions, generates DEV images, binaries, dependencies, and push to to Customer Manager.
5. **Create PR**: After testing, create a Pull Request (PR) to main.
6. **Activate QA Actions**: Once the PR is approved, QA actions are triggered, generates QA images, binaries, dependencies, and push to to Customer Manager.
7. **Merge to Main**: If all tests pass, merging to main activates RC actions and generates RC images, binaries, dependencies, and push to to Customer Manager.
8. **Release**: After testing in RC, create a new release to activate release actions.

## Hotfix
1. **Create Branch**: Start a new branch from main for the hotfix.
2. **Develop**: All developers related to the hotfix work on this branch.
3. **Create PR**: Once the hotfix is complete, create a PR to main.
4. **Activate RC Actions**: After merging the PR, RC actions are triggered, and RC images, binaries and dependencies are generated and pushed to to Customer Manager.
5. **Release**: If everything is satisfactory, create a new release to activate release actions.


## Actions
**DEV**->on push ***feature/\*\*/\*\**** branch<br>
**QA**->on pull-request: approved main branch<br>
**RC**->on push: main branch<br>
**RELEASE**-> on release<br>

## Service Directory Structure and Versioniong Guidelines
To ensure consistency and streamline our development workflow, please adhere to the following guidelines for organizing service directories, documentation, and versioning:

### Folder Naming Conventions
- Hyphens (-) in Folder Names alloed only for Docker Image Services

### Required Documentation for Each Service
Each service must include the following mandatory documentation files to ensure clarity and maintainability:
- README.md: Provides an overview of the service, including setup instructions, usage, and other relevant information.
- CHANGELOG.md: Logs all significant changes, such as new features, bug fixes, and improvements.
- scripts.json: Contains commands to run after update process, facilitating automation and deployment processes.

### Additional Files for Non-Image Services
For services that are not Docker images, an additional configuration file is required:

- files.json: Specifies the files associated with the service, including details necessary for compilation and deployment.
Example Structure for a file.json:
```
[
  {
    "name": "service-binary.exe",
    "is_binary": true,
    "destination_path": "bin/",
    "os": "windows",
    "arch": "amd64",
    "replace_previous": false
  },
  {
    "name": "config.yaml",
    "is_binary": false,
    "destination_path": "config/",
    "os": "linux",
    "arch": "amd64",
    "replace_previous": true
  }
]
```

### Versioning and Triggering Actions
Proper versioning is crucial for tracking changes and triggering automated workflows:
#### Update versions.json
- **When to Update**: Every time a change is made to any service.
- **Purpose**: Updating the version in versions.json serves as a trigger for GitHub Actions, initiating automated processes such as building and deploying services.