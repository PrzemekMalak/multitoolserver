#!/bin/sh
# Entrypoint script to run both the Go server and process monitor

# Start the process monitor in the background
python3 /monitor_processes.py &
MONITOR_PID=$!

# Start the Go server in the foreground
exec /bin/serv

# If the server exits, kill the monitor
kill $MONITOR_PID 2>/dev/null

