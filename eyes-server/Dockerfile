FROM debian:buster-slim

# RUN apt-get upgrade comment in for hadolint lint to fail
RUN useradd --home-dir /app --create-home --shell /bin/bash app

USER app
COPY . /app

ENTRYPOINT ["/app/entrypoint.bash"]

EXPOSE 8080
EXPOSE 7070
