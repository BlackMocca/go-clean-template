FROM postgres:11

RUN apt-get update && apt-get install -y wget

RUN wget https://github.com/BlackMocca/migrate/releases/download/v5.0/migrate.linux-amd64
RUN chmod +x migrate.linux-amd64
RUN mv migrate.linux-amd64 /usr/local/bin/migrate

ADD ./init /docker-entrypoint-initdb.d/

