use std::{net, sync::Arc};

use actix_web::{get, post, web, App, HttpServer, Responder};
use bytes::Bytes;
use dashmap::DashMap;
use openssl::ssl::{SslAcceptor, SslFiletype, SslMethod};
use opentelemetry::global;
use serde::{Deserialize, Serialize};
use tokio::sync::{mpsc, oneshot};
use tracing::{info, warn};
use tracing::{metadata::LevelFilter, span, Level, Span};
use tracing_subscriber::{fmt, layer::SubscriberExt, util::SubscriberInitExt};
use uuid::Uuid;

use crate::{
    api::{coordinator_get, coordinator_put, end_agent, get_self_ip, update_agent, EndReq},
    mq::{Command, MqClient, MqServer},
    usercode::Usercode,
};

enum Subscriber {
    Local(mpsc::Sender<Bytes>),
    Remote(String),
}

#[derive(PartialEq, Clone, Debug, Hash, Eq)]
struct ChannelKey(String, String); // dag_id, key

pub struct AgentServer {
    mq: MqServer,
    usercode: Usercode,
    subscribers: Arc<Db>, // channel -> (subscribers, history)
    self_ip: String,
    dag_instances: Arc<DagInstances>, // dag_id -> workflow_span
    states: Arc<StateInstances>,      // key -> state_span
}

type Db = DashMap<ChannelKey, (Vec<Subscriber>, Vec<Bytes>)>;
type DagInstances = DashMap<String, Arc<Span>>;
type StateInstances = DashMap<ChannelKey, Arc<Span>>;

async fn handle_publish(
    subscribers: Arc<Db>,
    channel: ChannelKey,
    resp: oneshot::Sender<(mpsc::Sender<Bytes>, Arc<Span>)>,
    self_ip: String,
    state_span: Arc<Span>,
) {
    let (tx, mut rx) = mpsc::channel(MqClient::DEFAULT_CAPACITY);
    resp.send((tx, state_span.clone())).unwrap();
    let _entered = state_span.enter();
    let state_span_clone = state_span.clone();
    let res = coordinator_put(channel.1.clone(), channel.0.clone()).await;
    let ips = res.ips;
    let subs = ips.into_iter().map(|ip| Subscriber::Remote(ip)).collect();
    let empty_history = Vec::new();
    subscribers.insert(channel.clone(), (subs, empty_history));

    // Thread for serving published messages.
    tokio::spawn(async move {
        let span = span!(
            parent: &*state_span_clone,
            Level::INFO,
            "Propagate change.",
        );
        let _enter = span.enter();
        // When receiving messages from the producer, the agent broadcast the
        // message to all subscribers.
        while let Some(msg) = rx.recv().await {
            let sub_span = span!(Level::INFO, "propagate", len = msg.len());
            let _enter = sub_span.enter();
            if let Some(mut subs) = subscribers.get_mut(&channel) {
                let (subs, history) = subs.value_mut();
                history.push(msg.clone());
                for sub in subs {
                    match sub {
                        Subscriber::Local(tx) => {
                            tx.send(msg.clone()).await.unwrap();
                        }
                        Subscriber::Remote(ip) => {
                            // When the target ip is the same as self_ip, there
                            // is no need to send the message.
                            if *ip != self_ip {
                                update_agent(
                                    ip.clone(),
                                    channel.1.clone(),
                                    channel.0.clone(),
                                    msg.clone(),
                                )
                                .await;
                            }
                        }
                    }
                }
            }
        }
        info!("Local end and clear: {:?}", channel);
        if let Some(mut subs) = subscribers.get_mut(&channel) {
            for sub in subs.value().0.iter() {
                match sub {
                    Subscriber::Local(_tx) => {
                        // Do nothing
                    }
                    Subscriber::Remote(ip) => {
                        if *ip != self_ip {
                            end_agent(ip.clone(), channel.1.clone(), channel.0.clone()).await;
                        }
                    }
                }
            }
            subs.0.clear();
        }
    });
}

async fn handle_subscribe(
    subscribers: Arc<Db>,
    channel: ChannelKey,
    resp: oneshot::Sender<(mpsc::Receiver<Bytes>, Arc<Span>)>,
    parent: Arc<Span>,
) {
    let (tx, rx) = mpsc::channel(MqClient::DEFAULT_CAPACITY);
    resp.send((rx, parent)).unwrap();
    if let Some(mut subs) = subscribers.get_mut(&channel) {
        // Update history
        for ch in subs.value().1.iter() {
            tx.send(ch.clone()).await.unwrap();
        }
        let sub = Subscriber::Local(tx);
        subs.value_mut().0.push(sub);
    } else {
        let empty_history = Vec::new();
        let sub = Subscriber::Local(tx);
        subscribers.insert(channel.clone(), (vec![sub], empty_history));
    }
    coordinator_get(channel.1.clone(), channel.0.clone()).await;
}

fn get_or_create_state_span(
    dag_span: &Span,
    states: &StateInstances,
    channel_key: &ChannelKey,
) -> Arc<Span> {
    if let Some(state_span) = states.get(&channel_key) {
        state_span.clone()
    } else {
        let state_span = Arc::new(span!(
            parent: dag_span,
            Level::INFO,
            "State.",
            key = channel_key.1.clone()
        ));
        states.insert(channel_key.clone(), state_span.clone());
        state_span
    }
}

fn clean_dag(
    dag_instances: &DashMap<String, Arc<Span>>,
    subscribers: &Db,
    states: &StateInstances,
    dag_id: &String,
) {
    dag_instances.remove(dag_id);
    subscribers.retain(|key, _| key.0 != *dag_id);
    states.retain(|key, _| key.0 != *dag_id);
}

fn init_tracing() {
    global::set_text_map_propagator(opentelemetry_jaeger::Propagator::new());
    let tracer = opentelemetry_jaeger::new_pipeline()
        .with_agent_endpoint(("simplest-agent.jaeger", 6831))
        .with_service_name("chitu-workflow")
        .install_batch(opentelemetry::runtime::Tokio)
        .unwrap();

    let opentelemetry = tracing_opentelemetry::layer().with_tracer(tracer);
    let filter = LevelFilter::INFO;
    tracing_subscriber::registry()
        .with(opentelemetry)
        .with(fmt::Layer::default())
        .with(filter)
        .try_init()
        .unwrap();
}

pub async fn start_server(
    usercode: Usercode,
    addrs: impl net::ToSocketAddrs,
    ls_addrs: impl net::ToSocketAddrs,
) {
    init_tracing();

    let subscribers: Arc<Db> = Arc::new(DashMap::new());
    let self_ip = get_self_ip().await.unwrap_or_else(|_e| {
        warn!("Remote coordinator access failed. Start as local agent.");
        "127.0.0.1".to_string()
    });

    let dag_instances: Arc<DagInstances> = Arc::new(DashMap::new());
    let dag_instances_clone = dag_instances.clone();

    let self_ip_clone = self_ip.clone();
    let subscribers_clone = subscribers.clone();

    let states: Arc<StateInstances> = Arc::new(DashMap::new());
    let states_clone = states.clone();

    let mq_handler = move |cmd| {
        let self_ip = self_ip_clone.clone();
        let subscribers = subscribers_clone.clone();
        let dag_instances = dag_instances_clone.clone();
        let states = states_clone.clone();
        async move {
            match cmd {
                Command::Publish {
                    channel,
                    dag_id,
                    resp,
                } => {
                    if let Some(dag_span) = dag_instances.get(&dag_id) {
                        let channel_key = ChannelKey(dag_id, channel);
                        let state_span =
                            get_or_create_state_span(&*dag_span, &*states, &channel_key);
                        handle_publish(
                            subscribers,
                            channel_key.clone(),
                            resp,
                            self_ip.clone(),
                            state_span,
                        )
                        .await;
                    } else {
                        panic!("Cannot find dag: {} when publish.", dag_id);
                    }
                }
                Command::Subscribe {
                    channel,
                    dag_id,
                    resp,
                } => {
                    if let Some(parent_span) = dag_instances.get(&dag_id) {
                        let channel_key = ChannelKey(dag_id.clone(), channel.clone());
                        let parent =
                            get_or_create_state_span(&*parent_span, &*states, &channel_key);
                        handle_subscribe(subscribers, channel_key, resp, parent).await;
                    } else {
                        panic!("Cannot find dag: {} when subscribing.", dag_id);
                    }
                }
            }
        }
    };
    let mq = MqServer::new(mq_handler);
    let server = AgentServer {
        mq,
        usercode,
        subscribers,
        self_ip,
        dag_instances,
        states,
    };

    let mut builder = SslAcceptor::mozilla_intermediate(SslMethod::tls()).unwrap();
    builder
        .set_private_key_file("server.key", SslFiletype::PEM)
        .unwrap();
    builder.set_certificate_chain_file("server.crt").unwrap();
    let server = web::Data::new(server);
    let http_server = HttpServer::new(move || {
        App::new()
            .app_data(server.clone())
            .app_data(web::PayloadConfig::new(1024 * 1024 * 1024))
            .service(async_invoke)
            .service(get_notify)
            .service(update)
            .service(end)
            .service(health)
            .service(root_invoke)
    })
    .bind(addrs)
    .unwrap()
    .bind_openssl(ls_addrs, builder)
    .unwrap()
    .run()
    .await
    .unwrap();
    http_server
}

#[derive(Deserialize)]
#[serde(rename_all = "camelCase")]
struct InvokeReq {
    req: String,
    dag_id: String,
}

// TODO: Async-invoke now does not support injected tracing span
#[post("/async-invoke")]
async fn async_invoke(server: web::Data<AgentServer>, req: web::Json<InvokeReq>) -> impl Responder {
    info!("Async invoke");
    let dag_id = req.dag_id.clone();
    let args = req.req.clone();
    let client = server.mq.fork_client();
    // TODO: Inject tracing span
    let span = if let Some(parent_span) = server.dag_instances.get(&dag_id) {
        parent_span.clone()
    } else {
        let span = Arc::new(span!(
            Level::INFO,
            "Workflow execution (subfunction).",
            dag_id = dag_id
        ));
        server.dag_instances.insert(dag_id.clone(), span.clone());
        span
    };
    tokio::spawn(async move {
        let entered = span.enter();
        server.usercode.invoke(client, dag_id.clone(), args).await;
        drop(entered);
        clean_dag(
            &*server.dag_instances,
            &*server.subscribers,
            &*server.states,
            &dag_id,
        );
    });
    ""
}

#[post("/")]
async fn root_invoke(server: web::Data<AgentServer>, args: String) -> impl Responder {
    info!("Root invoke");
    let client = server.mq.fork_client();
    let dag_id = Uuid::new_v4().to_string();
    let span = Arc::new(span!(Level::INFO, "Workflow execution.", dag_id = dag_id));
    server.dag_instances.insert(dag_id.clone(), span.clone());
    let entered = span.enter();
    let res = server.usercode.invoke(client, dag_id.clone(), args).await;
    drop(entered);
    clean_dag(
        &*server.dag_instances,
        &*server.subscribers,
        &*server.states,
        &dag_id,
    );
    res
}

#[derive(Deserialize, Serialize)]
#[serde(rename_all = "camelCase")]
struct GetNotifyReq {
    key: String,
    dag_id: String,
    ip: String,
}

#[post("/get-notify")]
async fn get_notify(
    server: web::Data<AgentServer>,
    body: web::Json<GetNotifyReq>,
) -> impl Responder {
    info!("Get-notify: {} from {}", body.key, body.ip);
    let key = ChannelKey(body.dag_id.clone(), body.key.clone());
    let sub = Subscriber::Remote(body.ip.clone());

    // Update history
    if let Some(mut subs) = server.subscribers.get_mut(&key) {
        if server.self_ip != body.ip.clone() {
            let history = &subs.value().1;
            for ch in history {
                update_agent(
                    body.ip.clone(),
                    body.key.clone(),
                    body.dag_id.clone(),
                    ch.clone(),
                )
                .await;
            }
        }
        subs.value_mut().0.push(sub);
    } else {
        let empty_history = Vec::new();
        server.subscribers.insert(key, (vec![sub], empty_history));
    }
    ""
}

#[derive(Deserialize)]
#[serde(rename_all = "camelCase")]
struct UpdateQuery {
    key: String,
    dag_id: String,
}

/// Since each channel can only have one writer, when remote update happends,
/// the remote must be the writer. In this case, the subscribers on this agent
/// must be `Local` subscribers.
#[post("/update")]
async fn update(
    server: web::Data<AgentServer>,
    query: web::Query<UpdateQuery>,
    msg: web::Bytes,
) -> impl Responder {
    // TODO: subspan `Load change` of state
    let key = ChannelKey(query.dag_id.clone(), query.key.clone());
    let state_span = server.states.get(&key).unwrap().clone();
    let span = span!(parent: &*state_span, Level::INFO, "load", key = query.key, len = msg.len());
    let _entered = span.enter();
    if let Some(mut subs) = server.subscribers.get_mut(&key) {
        let (subs, history) = subs.value_mut();
        history.push(msg.clone());
        for sub in subs {
            match sub {
                Subscriber::Local(tx) => {
                    tx.send(msg.clone()).await.unwrap();
                }
                Subscriber::Remote(_) => panic!("Updating from remote to remote."),
            }
        }
    } else {
        panic!("Updating from remote to non-subscribers.")
    }
    ""
}

#[post("/end")]
async fn end(server: web::Data<AgentServer>, body: web::Json<EndReq>) -> impl Responder {
    info!("Remote end: {}", body.key);
    let key = ChannelKey(body.dag_id.clone(), body.key.clone());
    if let Some(subs) = server.subscribers.get(&key) {
        let subs = subs.value().0.iter();
        // Check if there is any remote subscriber
        for sub in subs {
            match sub {
                Subscriber::Local(_tx) => {
                    // Do nothing
                }
                Subscriber::Remote(_ip) => {
                    panic!("Ending from remote to remote.");
                }
            }
        }
    } else {
        panic!("Ending from remote to non-subscribers.");
    }
    // server.subscribers.remove(&key);
    ""
}

#[get("/_/health")]
async fn health() -> impl Responder {
    "OK"
}
