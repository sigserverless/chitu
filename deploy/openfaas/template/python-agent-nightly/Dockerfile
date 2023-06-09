FROM rust:1.69 as builder

WORKDIR /home/rust

COPY config /home/.cargo/config

COPY main/Cargo.toml ./main/Cargo.toml
COPY chitu/Cargo.toml ./chitu/Cargo.toml
COPY chitu/chitu/Cargo.toml ./chitu/chitu/Cargo.toml
COPY chitu/chitu-macros/Cargo.toml ./chitu/chitu-macros/Cargo.toml

RUN mkdir -p ./chitu/chitu/src
RUN touch ./chitu/chitu/src/lib.rs

RUN mkdir -p ./chitu/chitu-macros/src
RUN touch ./chitu/chitu-macros/src/lib.rs

RUN mkdir -p ./main/src
RUN echo "fn main() {println!(\"if you see this, the build broke\")}" > ./main/src/main.rs

ENV CARGO_TARGET_DIR=/home/rust/target
WORKDIR /home/rust/main

RUN cargo fetch 
RUN cargo build --release

# Copy all the sources
WORKDIR /home/rust

COPY chitu ./chitu 
COPY main ./main 

RUN find . -name "*.rs" -exec touch {} \;

WORKDIR /home/rust/main
RUN cargo build --release

# FROM debian:11.7-slim as runner
FROM python:3.10.10 as runner

RUN groupadd app && useradd -r -g app app
RUN mkdir -p /home/app \
    && chown app /home/app

ENV USER=app

WORKDIR /home/app

COPY function/requirements.txt .

# requirements
RUN pip3 install -r ./requirements.txt -i https://pypi.mirrors.ustc.edu.cn/simple -t .

COPY --from=builder /home/rust/target/release/main ./
COPY lib-python ./
COPY function/ .

COPY server.crt ./
COPY server.key ./

ENV RUST_LOG="info"

# HEALTHCHECK --interval=3s CMD [ -e /tmp/.lock ] || exit 1

CMD ["./main"]