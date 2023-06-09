use std::{collections::HashMap, fmt::Debug, future::Future};

pub trait Change: Clone + Debug {
    type Value: Clone + Debug;
    fn apply(self, v: &mut Self::Value);
}

pub type MultiChange<C> = Vec<C>;

impl<C: Change> Change for MultiChange<C> {
    type Value = C::Value;
    fn apply(self, v: &mut Self::Value) {
        for ch in self {
            ch.apply(v);
        }
    }
}

#[derive(Clone, Debug, Serialize, Deserialize)]
pub enum TrivialChange<T: Clone + Debug> {
    Nop,
    Replace(T),
}

impl<T: Clone + Debug> Change for TrivialChange<T> {
    type Value = T;
    fn apply(self, v: &mut Self::Value) {
        match self {
            TrivialChange::Nop => {}
            TrivialChange::Replace(val) => *v = val,
        }
    }
}

use serde::{Deserialize, Serialize};
use tokio::{sync::mpsc, task::JoinHandle};

pub struct Diff<T, C = TrivialChange<T>>
where
    T: Default + Clone + Debug + Send + Sync + 'static,
    C: Change<Value = T> + Send + Sync + 'static,
{
    listeners: Vec<JoinHandle<()>>,
    senders: Vec<mpsc::UnboundedSender<C>>,
    result: JoinHandle<T>,
}

impl<T, C> Diff<T, C>
where
    T: Default + Clone + Debug + Send + Sync + 'static,
    C: Change<Value = T> + Send + Sync + 'static,
{
    pub fn new() -> Self {
        let (tx, mut rx) = mpsc::unbounded_channel::<C>();
        let handle = tokio::spawn(async move {
            let mut value = T::default();
            while let Some(ch) = rx.recv().await {
                ch.apply(&mut value)
            }
            value
        });
        Diff {
            listeners: Vec::new(),
            senders: vec![tx],
            result: handle,
        }
    }

    pub fn change(&self, ch: C) {
        for tx in &self.senders {
            tx.send(ch.clone()).unwrap();
        }
    }

    pub async fn end(self) -> T {
        for tx in self.senders {
            drop(tx);
        }
        for handle in self.listeners {
            handle.await.unwrap();
        }
        self.result.await.unwrap()
    }

    pub fn onchange(&mut self, mut handler: impl FnMut(C) + Send + 'static) {
        let handle = self.onchange_state((), move |(), ch| handler(ch));
        self.listeners.push(handle);
    }

    pub fn onchange_async<Fut>(&mut self, mut handler: impl FnMut(C) -> Fut + Send + Sync + 'static)
    where
        Fut: Future<Output = ()> + Send + 'static,
    {
        let handle = self.onchange_state_async((), move |(), ch| handler(ch));
        self.listeners.push(handle);
    }

    pub fn onchange_state<R: Send + 'static>(
        &mut self,
        init_state: R,
        mut handler: impl FnMut(R, C) -> R + Send + 'static,
    ) -> tokio::task::JoinHandle<R> {
        let (tx, mut rx) = mpsc::unbounded_channel::<C>();
        let handle = tokio::spawn(async move {
            let mut state = init_state;
            while let Some(ch) = rx.recv().await {
                state = handler(state, ch);
            }
            state
        });
        self.senders.push(tx);
        handle
    }

    pub fn onchange_blocking_state<R: Send + 'static>(
        &mut self,
        init_state: R,
        mut handler: impl FnMut(R, C) -> R + Send + 'static,
    ) -> tokio::task::JoinHandle<R> {
        let (tx, mut rx) = mpsc::unbounded_channel::<C>();
        let handle = tokio::task::spawn_blocking(move || {
            let mut state = init_state;
            while let Some(ch) = rx.blocking_recv() {
                state = handler(state, ch);
            }
            state
        });
        self.senders.push(tx);
        handle
    }

    pub fn onchange_state_async<R: Send + 'static, Fut>(
        &mut self,
        init_state: R,
        mut handler: impl FnMut(R, C) -> Fut + Send + Sync + 'static,
    ) -> tokio::task::JoinHandle<R>
    where
        Fut: Future<Output = R> + Send + 'static,
    {
        let (tx, mut rx) = mpsc::unbounded_channel::<C>();
        let handle = tokio::spawn(async move {
            let mut state = init_state;
            while let Some(ch) = rx.recv().await {
                state = handler(state, ch).await;
            }
            state
        });
        self.senders.push(tx);
        handle
    }
}

impl<T> Diff<Vec<T>, VecChange<T>>
where
    T: Clone + Debug + Send + Sync + 'static,
{
    pub async fn push(&mut self, val: T) {
        self.change(VecChange::Push(val));
    }
}

#[derive(Clone, Debug, Serialize, Deserialize)]
pub enum VecChange<T> {
    Push(T),
}

impl<T: Clone + Debug> Change for VecChange<T> {
    type Value = Vec<T>;
    fn apply(self, v: &mut Self::Value) {
        match self {
            VecChange::Push(val) => v.push(val),
        }
    }
}

use std::hash::Hash;

#[derive(Clone, Debug, Serialize, Deserialize)]
pub enum DictChange<K, V>
where
    K: Hash + Eq + Debug,
{
    Set(K, V),
}

impl<K: Eq + Debug + Hash + Clone, V: Clone + Debug> Change for DictChange<K, V> {
    type Value = HashMap<K, V>;
    fn apply(self, hashmap: &mut Self::Value) {
        match self {
            Self::Set(k, v) => {
                hashmap.insert(k, v);
            }
        }
    }
}

impl<K, V> Diff<HashMap<K, V>, DictChange<K, V>>
where
    K: Hash + Eq + Debug + Clone + Send + Sync + 'static,
    V: Clone + Debug + Send + Sync + 'static,
{
    pub async fn set(&mut self, key: K, val: V) {
        self.change(DictChange::Set(key, val));
    }
}
