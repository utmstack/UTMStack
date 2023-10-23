FROM python:3.9 AS compile-image

# Keeps Python from generating .pyc files in the container
ENV PYTHONDONTWRITEBYTECODE=1

# Turns off buffering for easier container logging
ENV PYTHONUNBUFFERED=0

WORKDIR /app

RUN python -m venv /app/venv
ENV PATH="/app/venv/bin:$PATH"

# Install pip requirements
COPY requirements.txt .
RUN python -m pip install --no-cache-dir -r requirements.txt

FROM python:3.9-slim

WORKDIR /app

COPY . /app/mutate

COPY --from=compile-image /app/venv /app/venv

# Add PostgreSQL client library and curl
RUN apt-get update \
    && apt-get -y install libpq-dev gcc \
    && apt-get -y install curl


ENV PYTHONPATH = "${PYTHONPATH}:/app"
ENV PATH="/app/venv/bin:$PATH"

# Start the application
CMD ["python", "/app/mutate/main.py"]