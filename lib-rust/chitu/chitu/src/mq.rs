use std::{future::Future, sync::Arc};

use bytes::Bytes;
use tokio::sync::{mpsc, oneshot};
use tracing::Span;

type Responder<T> = oneshot::Sender<T>;

#[derive(Debug)]
pub enum Command {
    Publish {
        channel: String,
        dag_id: String,
        resp: Responder<(mpsc::Sender<Bytes>, Arc<Span>)>,
    },
    Subscribe {
        channel: String,
        dag_id: String,
        resp: Responder<(mpsc::Receiver<Bytes>, Arc<Span>)>,
    },
}

pub struct MqServer {
    tx: mpsc::Sender<Command>,
}

impl MqServer {
    pub fn new<Fut>(handler: impl Fn(Command) -> Fut + Send + Sync + 'static) -> MqServer
    where
        Fut: Future<Output = ()> + Send + Sync + 'static,
    {
        let (tx, mut rx) = mpsc::channel(32);
        tokio::spawn(async move {
            while let Some(cmd) = rx.recv().await {
                handler(cmd).await;
            }
        });
        MqServer { tx: tx }
    }

    pub fn fork_client(&self) -> MqClient {
        MqClient {
            tx: self.tx.clone(),
        }
    }
}

pub struct MqClient {
    tx: mpsc::Sender<Command>,
}

impl MqClient {
    pub const DEFAULT_CAPACITY: usize = 32;

    pub async fn publish(&self, channel: String, dag_id: String) -> (mpsc::Sender<Bytes>, Arc<Span>) {
        let (tx, rx) = oneshot::channel();
        let cmd = Command::Publish {
            channel: channel,
            dag_id: dag_id,
            resp: tx,
        };
        self.tx.send(cmd).await.unwrap();
        rx.await.unwrap()
    }

    pub async fn subscribe(&self, channel: String, dag_id: String) -> (mpsc::Receiver<Bytes>, Arc<Span>) {
        let (tx, rx) = oneshot::channel();
        let cmd = Command::Subscribe {
            channel: channel,
            dag_id: dag_id,
            resp: tx,
        };
        self.tx.send(cmd).await.unwrap();
        rx.await.unwrap()
    }
}

impl Clone for MqClient {
    fn clone(&self) -> Self {
        MqClient {
            tx: self.tx.clone(),
        }
    }
}
