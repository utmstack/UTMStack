# Mutate

The mutate service facilitates the creation of Logstash pipelines. This README outlines the necessary environment variables, volume mappings, and instructions for getting the application running via Docker.

## Environment Variables

- `DB_HOST`: The hostname of the local PostgreSQL database server.
- `DB_PORT`: The port number on which the PostgreSQL database server is listening.
- `DB_NAME`: The name of the PostgreSQL database to connect to.
- `DB_USER`: The username to authenticate with the PostgreSQL database server.
- `DB_PASSWORD`: The password to authenticate with the PostgreSQL database server.
- `ENCRYPTION_KEY`: Key used for encryption within UTMstack.
- `SERVER_NAME`

## Volumes

- Host Volume: `/path/on/host/to/logstash/pipelines` -> Container Path: `/usr/share/logstash/pipelines`
- Host Volume: `/path/on/host/to/logstash/config/pipelines.yml` -> Container Path: `/usr/share/logstash/config/pipelines.yml`

## Getting Started

1. **Clone the Repository**: If not already done, clone the repository to your local machine:
   ```bash
   git clone [your-repository-url] && cd [your-repository-dir]
    ```
   
2. **Build the Docker Image for the `mutate` Module**:
   ```bash
   docker build -t mutate-module:latest .
   ```

3. **Run the Docker Container**:
Before running the container, ensure you have set the required environment variables. You can use a .env file or pass them directly in the command line.

    ```bash
       docker run -d \
        -e DB_HOST=your_db_host \
        -e DB_PORT=your_db_port \
        -e DB_NAME=your_db_name \
        -e DB_USER=your_db_user \
        -e DB_PASSWORD=your_db_password \
        -e ENCRYPTION_KEY=your_encryption_key \
        -v /path/on/host/to/logstash/pipelines:/usr/share/logstash/pipelines \
        -v /path/on/host/to/logstash/config/pipelines.yml:/usr/share/logstash/config/pipelines.yml \
        mutate:latest
   ```
