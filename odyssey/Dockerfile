FROM ubuntu:22.04 AS builder

# Установка зависимостей
RUN apt-get update && apt-get install -y \
    build-essential \
    cmake \
    git \
    libpq-dev \
    libssl-dev \
    postgresql-client

# Клонирование и сборка Odyssey
RUN git clone --branch 1.4rc https://github.com/yandex/odyssey.git /odyssey \
    && cd /odyssey \
    && make local_build

# Финальный образ
FROM ubuntu:22.04

RUN apt-get update && apt-get install -y \
    libssl3 \
    postgresql-client \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /odyssey/build/sources/odyssey /usr/bin/odyssey

CMD ["odyssey", "/etc/odyssey/odyssey.conf"]