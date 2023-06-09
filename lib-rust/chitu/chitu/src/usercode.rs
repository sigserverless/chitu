use std::collections::HashMap;
use std::path::Path;
use std::process::Stdio;

use bytes::Bytes;
use tracing::info;
use nix::{sys::stat::Mode, unistd::mkfifo};
use serde::{Deserialize, Serialize};
use tokio::fs::{self, OpenOptions};
use tokio::io::{AsyncBufReadExt, AsyncReadExt, AsyncWriteExt};
use tokio::process::Command;
use tokio::sync::mpsc;

use crate::{mq::MqClient, stub::AgentStub};

pub enum Usercode {
    Native(fn(AgentStub, String) -> String),
    Runner(&'static str, &'static [&'static str]),
}

impl Usercode {
    pub async fn invoke(&self, client: MqClient, dag_id: String, args: String) -> String {
        match self {
            Usercode::Native(f) => {
                let stub = AgentStub::new(client, dag_id);
                f(stub, args)
            }
            Usercode::Runner(cmd, cmd_args) => {
                start_runner(
                    "write.fifo",
                    "read.fifo",
                    dag_id,
                    client,
                    cmd,
                    cmd_args,
                    args,
                )
                .await
            }
        }
    }
}

#[derive(Deserialize)]
#[serde(rename_all = "camelCase")]
#[serde(tag = "kind")]
enum CommandHelper {
    Import { key: String },
    Export { key: String },
    Pub { key: String, len: i64 },
}

#[derive(Serialize)]
struct SubHelper {
    key: String,
    len: i64,
}

#[derive(Debug)]
struct Sub {
    key: String,
    msg: Bytes,
}

async fn start_runner(
    write_fifo_path: &'static str,
    read_fifo_path: &'static str,
    dag_id: String,
    client: MqClient,
    cmd: &'static str,
    cmd_args: &'static [&'static str],
    input: String,
) -> String {
    if !Path::new(read_fifo_path).exists() {
        mkfifo(read_fifo_path, Mode::S_IRWXU).unwrap();
    }
    if !Path::new(write_fifo_path).exists() {
        mkfifo(write_fifo_path, Mode::S_IRWXU).unwrap();
    }

    let (sub_tx, mut sub_rx) = mpsc::channel(MqClient::DEFAULT_CAPACITY);
    let dag_id_clone = dag_id.clone();

    let sub_process = tokio::spawn(async move {
        let mut child = Command::new(cmd)
            .args(cmd_args)
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .env("DAG_ID", dag_id)
            .spawn()
            .unwrap();

        info!("Child process executed");

        if let Some(mut stdin) = child.stdin.take() {
            stdin.write_all(input.as_bytes()).await.unwrap();
        }

        info!("Child process input written");

        while let Some(output) = child.stdout.take() {
            let mut reader = tokio::io::BufReader::new(output);
            let mut buffer = String::new();
            while reader.read_line(&mut buffer).await.unwrap() > 0 {
                info!("[handler] {}", buffer.trim_end());
                buffer.clear();
            }
        }

        info!("Child process output taken");
    });

    let serve_usercode = tokio::spawn(async move {
        let write_fifo = OpenOptions::new()
            .read(true)
            .open(write_fifo_path)
            .await
            .unwrap();
        let mut write_fifo = tokio::io::BufReader::new(write_fifo);
        let mut buffer = String::new();

        let mut exports = HashMap::new();
        while write_fifo.read_line(&mut buffer).await.unwrap() > 0 {
            // TODO: tracing
            let command = serde_json::from_str::<CommandHelper>(buffer.trim_end()).unwrap();
            match command {
                CommandHelper::Import { key } => {
                    let (mut pub_rx, _span) = client.subscribe(key.clone(), dag_id_clone.clone()).await;
                    let tx = sub_tx.clone();
                    tokio::spawn(async move {
                        while let Some(msg) = pub_rx.recv().await {
                            let sub_cmd = Sub {
                                key: key.clone(),
                                msg: msg,
                            };
                            tx.send(sub_cmd).await.unwrap();
                        }
                    });
                }
                CommandHelper::Export { key } => {
                    if exports.contains_key(&key) {
                        panic!("Cannot export the same key twice: {}", key);
                    }
                    let tx = client.publish(key.clone(), dag_id_clone.clone()).await;
                    exports.insert(key, tx);
                }
                CommandHelper::Pub { key, len } => {
                    if let Some((tx, _span)) = exports.get(&key) {
                        let mut msg = vec![0u8; len as usize];
                        write_fifo.read_exact(&mut msg).await.unwrap();
                        tx.send(msg.into()).await.unwrap();
                    } else {
                        panic!("Pub before export: {}", key);
                    }
                }
            }
            buffer.clear();
        }
        info!("Serve usercode done");
    });

    let fifo_writer = tokio::spawn(async move {
        let mut read_fifo = OpenOptions::new()
            .write(true)
            .open(read_fifo_path)
            .await
            .unwrap();
        while let Some(sub_cmd) = sub_rx.recv().await {
            let sub_cmd_helper = SubHelper {
                key: sub_cmd.key,
                len: sub_cmd.msg.len() as i64,
            };
            let sub_cmd_helper_str = serde_json::to_string(&sub_cmd_helper).unwrap();
            read_fifo
                .write_all(sub_cmd_helper_str.as_bytes())
                .await
                .unwrap();
            read_fifo.write_all(b"\n").await.unwrap();
            read_fifo.write_all(&sub_cmd.msg).await.unwrap();
        }
        info!("Fifo writer done");
    });

    sub_process.await.unwrap();
    serve_usercode.await.unwrap();
    fifo_writer.await.unwrap();

    fs::remove_file(write_fifo_path).await.unwrap();
    fs::remove_file(read_fifo_path).await.unwrap();

    "Handler as subprocess executed. Please check the logs for handler output.".to_string()
}
