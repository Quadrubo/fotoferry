#!/bin/sh
if [ -z "$CRON_SCHEDULE" ]; then
  export CRON_SCHEDULE='0 * * * *'
fi

echo "CRON_SCHEDULE: $CRON_SCHEDULE"

envsubst < /app/crontab.template | crontab
