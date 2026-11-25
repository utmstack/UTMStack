FROM python:3.12-slim

WORKDIR /app

COPY pyproject.toml ./
COPY src ./src

RUN pip install --no-cache-dir .

# The chat and the manual triage are served here; the alerts come from NATS.
EXPOSE 8090

HEALTHCHECK --start-period=20s --interval=30s --timeout=10s --retries=3 \
  CMD python -c "import urllib.request,sys; sys.exit(0 if urllib.request.urlopen('http://localhost:8090/health', timeout=5).status == 200 else 1)"

CMD ["python", "-m", "socai"]
