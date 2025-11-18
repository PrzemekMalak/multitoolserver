FROM --platform=linux/amd64 golang:1.24-bullseye AS build

RUN mkdir src/multitoolserver
WORKDIR /src/multitoolserver
COPY ./src .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/serv

# Python build stage for monitoring script
FROM --platform=linux/amd64 python:3.11-slim AS python-base
COPY requirements-monitor.txt /tmp/requirements-monitor.txt
RUN pip install --no-cache-dir -r /tmp/requirements-monitor.txt

# Final stage
FROM --platform=linux/amd64 python:3.11-slim
# Copy Python dependencies from python-base
COPY --from=python-base /usr/local/lib/python3.11/site-packages /usr/local/lib/python3.11/site-packages
COPY --from=python-base /usr/local/bin /usr/local/bin

# Copy Go binary
COPY --from=build /bin/serv /bin/serv

# Copy monitoring script
COPY monitor_processes.py /monitor_processes.py

# Copy and set up entrypoint
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["/docker-entrypoint.sh"]