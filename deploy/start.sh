#!/bin/sh
# Start Go backend in background
/server &

# Start nginx in foreground
nginx -g 'daemon off;'
