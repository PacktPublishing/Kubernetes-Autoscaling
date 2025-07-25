FROM grafana/k6:latest

COPY test-image.png /data/test-image.png
COPY k6-load-test.js /scripts/k6-load-test.js

USER root
RUN chown -R 12345:12345 /data /scripts
USER 12345

WORKDIR /scripts