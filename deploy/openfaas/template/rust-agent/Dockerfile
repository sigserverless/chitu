FROM rust:1.69 as builder

WORKDIR /home/rust

COPY config /home/.cargo/config

# Try to build once, downloading dependencies and save as cache
COPY ./default-dep.toml ./function/Cargo.toml 
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

RUN mkdir -p ./function/src
RUN touch ./function/src/lib.rs

ENV CARGO_TARGET_DIR=/home/rust/target
WORKDIR /home/rust/main

RUN cargo fetch 
RUN cargo build --release

COPY function/Cargo.toml ./function/Cargo.toml
RUN cargo fetch 

# Copy all the sources
WORKDIR /home/rust

COPY function ./function
COPY chitu ./chitu
COPY main ./main

RUN find . -name "*.rs" -exec touch {} \;

WORKDIR /home/rust/main
RUN cargo build --release

# FROM alpine:3.17.3 as runner
FROM debian:11.7-slim as runner

# RUN addgroup -S app && adduser -S -g app app
RUN groupadd app && useradd -r -g app app
RUN mkdir -p /home/app \
    && chown app /home/app

ENV USER=app

WORKDIR /home/app

COPY server.crt ./
COPY server.key ./

COPY --from=builder /home/rust/target/release/main ./

ENV RUST_LOG="info"

CMD ["./main"]