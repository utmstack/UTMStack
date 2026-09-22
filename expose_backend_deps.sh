#!/bin/bash

docker service update --publish-add 5432:5432  utmstack_postgres &
docker service update --publish-add 9009:9000  utmstack_clickhouse &


