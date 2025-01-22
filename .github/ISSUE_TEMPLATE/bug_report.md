---
name: Bug Report
about: File a bug report
title: "[BUG] <short description of the bug>"
labels: [bug]
assignees: ''

body:
  - type: markdown
    attributes:
      value: |
        To ensure efficient bug tracking, kindly confirm that the issue you're experiencing is indeed a bug. For general questions, configuration support, or discussions, please refer to our GitHub Discussions.

        We appreciate you taking the time to report this bug!
  - type: textarea
    id: what-happened
    attributes:
      label: What happened?
      description: Also tell us, what did you expect to happen?
      placeholder: Tell us what you see!
      value: "A bug happened!"
    validations:
      required: true
  - type: dropdown
    id: version
    attributes:
      label: Version
      description: What version of our software are you running?
      options:
        - v10
        - v11-preview
    validations:
      required: true
  - type: textarea
    id: steps-to-reproduce
    attributes:
      label: Steps to reproduce
      description: How can we reproduce the bug?
      placeholder: |
        1. Go to '...'
        2. Click on '....'
        3. Scroll down to '....'
        4. See error
    validations:
      required: true
  - type: textarea
    id: code-examples
    attributes:
      label: Code examples
      description: |
        If possible, please provide a brief code example or a link to a
        repository that can be used to reproduce the bug.
  - type: textarea
    id: logs
    attributes:
      label: Relevant logs and/or screenshots
      description: Please copy and paste any relevant log output or screenshots here.

---
