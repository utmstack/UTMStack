FROM python:3.11

WORKDIR /app

COPY . /app

# Keeps Python from generating .pyc files in the container
ENV PYTHONDONTWRITEBYTECODE=1

# Turns off buffering for easier container logging
ENV PYTHONUNBUFFERED=0

ENV PATH="/app/venv/bin:$PATH"

ENV PYTHONPATH = "${PYTHONPATH}:/app"

RUN apt-get update \
    && apt-get -y install libpq-dev gcc \
    && apt-get -y install curl

RUN python -m venv /app/venv

RUN python -m pip install --no-cache-dir -r /app/requirements.txt

# Start the application
CMD ["python", "main.py"]