use serde::de::DeserializeOwned;
use tracing::Instrument;
use std::fmt::Debug;
use std::sync::Arc;
use tokio::task::JoinHandle;
use tracing::span;
use tracing::Level;

use serde::Serialize;

use crate::api::coordinator_trigger;
use crate::diff::*;
use crate::mq::*;

pub struct AgentStub {
    mq: MqClient,
    dag_id: String,
}

impl AgentStub {
    pub fn new(mq: MqClient, dag_id: String) -> AgentStub {
        AgentStub { mq, dag_id }
    }

    pub async fn export<T, C>(&self, key: &str, diff: &mut Diff<T, C>)
    where
        T: Default + Clone + Debug + Send + Sync + 'static,
        C: Change<Value = T> + Serialize + Send + Sync + 'static,
    {
        let mq = self.mq.clone();
        let dag_id = self.dag_id.clone();

        let (tx, parent) = mq.publish(key.to_string(), dag_id).await;

        let ser_span = Arc::new(span!(parent: &*parent, Level::INFO, "Serialize change."));
        let _entered = ser_span.enter();
        let ser_span_clone = ser_span.clone();

        diff.onchange_async(move |ch| {
            let tx = tx.clone();
            let span = span!(parent: &*ser_span_clone, Level::INFO, "serialize");
            async move {
                let msg = bincode::serialize(&ch).unwrap();
                tx.send(msg.into()).await.unwrap();
            }.instrument(span)
        });
    }

    pub async fn import<T, C>(&self, key: &str, diff: Diff<T, C>) -> JoinHandle<Diff<T, C>>
    where
        T: Default + Clone + Debug + Send + Sync + 'static,
        C: Change<Value = T> + DeserializeOwned + Send + Sync + 'static,
    {
        let mq = self.mq.clone();
        let key = key.to_string();
        let dag_id = self.dag_id.clone();
        let (mut rx, parent) = mq.subscribe(key, dag_id).await;

        // let comp_span = Arc::new(span!(parent: &*parent, Level::INFO, "Apply change (compute)."));
        // let _entered = comp_span.enter();

        let deser_span = span!(parent: &*parent, Level::INFO, "Deserialize change.");
        let deser_span_clone = deser_span.clone();
        let _entered = deser_span.enter();
        
        let handle = tokio::spawn(async move {
            while let Some(b) = rx.recv().await {
                let span = span!(parent: &deser_span_clone, Level::INFO, "deserialize");
                let _entered = span.enter();
                let ch = bincode::deserialize(b.as_ref()).unwrap();
                diff.change(ch);
            }
            diff
        });
        handle
    }

    pub async fn trigger(&self, fname: &str, args: String) {
        coordinator_trigger(self.dag_id.clone(), fname.to_string(), args).await
    }
}
