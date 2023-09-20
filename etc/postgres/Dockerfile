FROM postgres:13

HEALTHCHECK --start-period=120s --interval=60s --timeout=60s \
  CMD pg_isready -U postgres || exit 1
