# The deployment Image
FROM docker.io/oraclelinux:8-slim

# Copy the native executable into the containers
COPY user-auditor .
ENTRYPOINT ["./user-auditor"]
