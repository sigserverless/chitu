use chitu::stub::AgentStub;
use chitu::diff::{Change, Diff};
use chitu_macros::async_handler;
use chrono::prelude::*;
use chrono::DateTime;
use chrono::Duration;
use chrono::Local;
use rand::Rng;
use serde::{Deserialize, Serialize};
use std::fmt::Debug;
use tracing::{span, Level};
use tracing::Instrument;

#[derive(Debug, Clone, Deserialize, Serialize)]
struct UserVisit {
    ip: String,
    dest: String,
    time: i64,
    revenue: f64,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
struct Ranking {
    url: String,
    rank: f64,
}

#[derive(Debug, Clone)]
struct ResultValue {
    total_revenue: f64,
    avg_rank: f64,
    occurence: i32,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
struct Result {
    ip: String,
    total_revenue: f64,
    avg_rank: f64,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
struct JoinResult {
    ip: String,
    revenue: f64,
    rank: f64,
}

#[derive(Deserialize, Serialize, Debug, Clone)]
struct VecExtend<T>(Vec<T>);

impl<T: Clone + Debug> Change for VecExtend<T> {
    type Value = Vec<T>;
    fn apply(self, _v: &mut Self::Value) {
        // v.extend(self.0);
    }
}

use std::collections::HashMap;

fn gen_random_ip() -> String {
    let ip = vec![
        rand::thread_rng().gen_range(0..255).to_string(),
        rand::thread_rng().gen_range(0..255).to_string(),
        rand::thread_rng().gen_range(0..255).to_string(),
        rand::thread_rng().gen_range(0..255).to_string(),
    ];
    ip.join(".")
}

fn gen_random_url() -> String {
    let charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";
    let mut url = String::from("http://");
    for _ in 0..10 {
        url.push(
            charset
                .chars()
                .nth(rand::thread_rng().gen_range(0..charset.len()))
                .unwrap(),
        );
    }
    url.push_str(".com");
    url
}

fn gen_random_time(start: DateTime<Local>, end: DateTime<Local>) -> DateTime<Local> {
    let duration = end - start;
    let secs = duration.num_seconds();
    let rand_secs = rand::thread_rng().gen_range(0..secs);
    start + Duration::seconds(rand_secs)
}

fn gen_user_visits(n: i32, groups: i32, urls: &Vec<String>) -> Vec<UserVisit> {
    let mut ips = Vec::with_capacity(groups as usize);
    for _ in 0..groups {
        ips.push(gen_random_ip());
    }

    let mut user_visits = Vec::with_capacity(n as usize);
    let start = Local.with_ymd_and_hms(1970, 1, 1, 0, 0, 0).unwrap();
    let end = Local.with_ymd_and_hms(2020, 1, 1, 0, 0, 0).unwrap();
    for _ in 0..n {
        user_visits.push(UserVisit {
            ip: ips[rand::thread_rng().gen_range(0..groups) as usize].clone(),
            dest: urls[rand::thread_rng().gen_range(0..urls.len())].clone(),
            time: gen_random_time(start, end).timestamp(),
            revenue: rand::thread_rng().gen_range(0.0..100.0),
        })
    }
    user_visits
}

fn gen_rankings(n: i32) -> (Vec<Ranking>, Vec<String>) {
    let mut rankings = Vec::with_capacity(n as usize);
    let mut urls = Vec::with_capacity(n as usize);
    for _ in 0..n {
        let url = gen_random_url();
        urls.push(url.clone());
        rankings.push(Ranking {
            url,
            rank: rand::thread_rng().gen_range(0.0..100.0),
        })
    }
    (rankings, urls)
}

fn sort_by_total_revenue(rs: &mut Vec<Result>) {
    rs.sort_by(|a, b| a.total_revenue.partial_cmp(&b.total_revenue).unwrap());
}

use std::sync::Arc;

async fn chitu_running(uvs: Vec<UserVisit>, ranks: Vec<Ranking>, chunk_size: usize) {
    let mode = if chunk_size == uvs.len() {
        "local"
    } else {
        "chitu"
    };
    let workflow_span = Arc::new(span!(Level::INFO, "Q3 benchmark", mode = mode));
    let _workflow_guard = workflow_span.enter();

    // init diffs
    let mut duvs = Diff::new();
    let mut filtered_uvs = Diff::new();
    let mut joined_results = Diff::new();

    let mut rank_map = HashMap::new();
    for rank in ranks {
        rank_map.insert(rank.url, rank.rank);
    }
    let rank_map = Arc::new(rank_map);

    let workflow_span_clone = workflow_span.clone();
    // subscribe diffs
    let grouped_results = joined_results.onchange_state_async(
        HashMap::new(),
        move |mut grouped_results: HashMap<String, ResultValue>, ch: VecExtend<JoinResult>| {
            let span = span!(parent: &*workflow_span_clone, Level::INFO, "grouping", len = ch.0.len());
            async move {
                for r in ch.0 {
                    let mut entry = grouped_results.entry(r.ip).or_insert(ResultValue {
                        total_revenue: 0.0,
                        avg_rank: 0.0,
                        occurence: 0,
                    });
                    entry.avg_rank += r.rank;
                    entry.total_revenue += r.revenue;
                    entry.occurence += 1;
                }
                grouped_results
            }.instrument(span)
        },
    );

    let workflow_span_clone = workflow_span.clone();
    let joined_results = filtered_uvs.onchange_state_async(
        joined_results,
        move |joined_results, ch: VecExtend<UserVisit>| {
            let span = span!(parent: &*workflow_span_clone, Level::INFO, "joining", len = ch.0.len());
            let rank_map = rank_map.clone();
            async move {
                let mut joined_ch = vec![];
                for uv in ch.0 {
                    let rank = rank_map.get(&uv.dest).unwrap();
                    joined_ch.push(JoinResult {
                        ip: uv.ip,
                        revenue: uv.revenue,
                        rank: *rank,
                    });
                }
                joined_results.change(VecExtend(joined_ch));
                joined_results
            }.instrument(span)
        },
    );

    let lower_time = Local.with_ymd_and_hms(1975, 1, 1, 0, 0, 0).unwrap().timestamp();
    let upper_time = Local.with_ymd_and_hms(2015, 1, 1, 0, 0, 0).unwrap().timestamp();

    let workflow_span_clone = workflow_span.clone();
    let filtered_uvs = duvs.onchange_state_async(
        filtered_uvs,
        move |filtered_uvs, ch: VecExtend<UserVisit>| {
            let span = span!(parent: &*workflow_span_clone, Level::INFO, "filtering", len = ch.0.len());
            async move {
                let res = ch
                    .0
                    .into_iter()
                    .filter(|uv| uv.time < upper_time && uv.time > lower_time)
                    .collect::<Vec<_>>();
                filtered_uvs.change(VecExtend(res));
                filtered_uvs
            }.instrument(span)
        },
    );

    let uvs_chunks = uvs.chunks(chunk_size as usize);
    for uvs_chunk in uvs_chunks {
        duvs.change(VecExtend(uvs_chunk.to_vec()));
    }

    // synchronization
    duvs.end().await;
    filtered_uvs.await.unwrap().end().await;
    joined_results.await.unwrap().end().await;

    // get results
    let mut grouped_results = grouped_results.await.unwrap();

    let sort_span = span!(parent: &*workflow_span, Level::INFO, "sorting");
    let _entered = sort_span.enter();
    
    for mut result in &mut grouped_results {
        result.1.avg_rank /= result.1.occurence as f64;
    }

    let mut sorted_results: Vec<_> = grouped_results
        .into_iter()
        .map(|r| Result {
            ip: r.0,
            total_revenue: r.1.total_revenue,
            avg_rank: r.1.avg_rank,
        })
        .collect();
    sort_by_total_revenue(&mut sorted_results);
    println!("Last one: ip: {}, avg_rank: {}", sorted_results[0].ip, sorted_results[0].avg_rank);
}

async fn redis_running(uvs: Vec<UserVisit>, ranks: Vec<Ranking>) {
    // use futures::prelude::*;
    use redis::AsyncCommands;

    // uncache 
    let uvs_chunks = uvs.chunks(uvs.len() as usize);
    let mut uvs = Vec::new();
    for uvs_chunk in uvs_chunks {
        uvs.extend(uvs_chunk.to_vec());
    }

    let client = redis::Client::open("redis://redis.openfaas:6379").unwrap();
    let mut redis = client.get_async_connection().await.unwrap();

    let workflow_span = span!(Level::INFO, "Q3 benchmark", mode = "redis");
    let _workflow_guard = workflow_span.enter();

    let mut rank_map = HashMap::new();
    for rank in ranks {
        rank_map.insert(rank.url, rank.rank);
    }

    let lower_time: DateTime<Local> = Local.with_ymd_and_hms(1975, 1, 1, 0, 0, 0).unwrap();
    let upper_time: DateTime<Local> = Local.with_ymd_and_hms(2015, 1, 1, 0, 0, 0).unwrap();

    // filter
    let filtered_uvs = span!(parent: &workflow_span, Level::INFO, "filtering").in_scope(|| {
        uvs
            .into_iter()
            .filter(|uv| uv.time < upper_time.timestamp() && uv.time > lower_time.timestamp())
            .collect::<Vec<_>>()
    });

    // trans
    let msg = span!(parent: &workflow_span, Level::INFO, "serializing").in_scope(|| 
        bincode::serialize(&filtered_uvs).unwrap()
    );
    let _: () = redis.set("filtered_uvs", &msg).instrument(span!(parent: &workflow_span, Level::INFO, "saving")).await.unwrap();
    let msg: Vec<u8> = redis.get("filtered_uvs").instrument(span!(parent: &workflow_span, Level::INFO, "loading")).await.unwrap();
    let filtered_uvs: Vec<UserVisit> = span!(parent: &workflow_span, Level::INFO, "deserializing").in_scope(|| 
        bincode::deserialize(&msg).unwrap()
    );

    // join
    let joined_results = span!(parent: &workflow_span, Level::INFO, "joining").in_scope(|| {
        filtered_uvs
            .into_iter()
            .map(|uv| {
                let rank = rank_map.get(&uv.dest).unwrap();
                JoinResult {
                    ip: uv.ip,
                    revenue: uv.revenue,
                    rank: *rank,
                }
            })
            .collect::<Vec<_>>()
    });

    // trans
    let msg = span!(parent: &workflow_span, Level::INFO, "serializing").in_scope(|| 
        bincode::serialize(&joined_results).unwrap()
    );
    let _: () = redis.set("joined_results", &msg).instrument(span!(parent: &workflow_span, Level::INFO, "saving")).await.unwrap();
    let msg: Vec<u8> = redis.get("joined_results").instrument(span!(parent: &workflow_span, Level::INFO, "loading")).await.unwrap();
    let joined_results: Vec<JoinResult> = span!(parent: &workflow_span, Level::INFO, "deserializing").in_scope(|| 
        bincode::deserialize(&msg).unwrap()
    );

    // group
    let grouped_results = span!(parent: &workflow_span, Level::INFO, "grouping").in_scope(|| {
        joined_results
            .into_iter()
            .fold(HashMap::new(), |mut map, jr| {
                let entry = map.entry(jr.ip).or_insert(ResultValue {
                    total_revenue: 0.0,
                    avg_rank: 0.0,
                    occurence: 0,
                });
                entry.total_revenue += jr.revenue;
                entry.avg_rank += jr.rank;
                entry.occurence += 1;
                map
            })
            .into_iter()
            .map(|(ip, gr)| Result {
                ip,
                total_revenue: gr.total_revenue,
                avg_rank: gr.avg_rank / gr.occurence as f64,
            })
            .collect::<Vec<_>>()
    });

    // trans
    let msg = span!(parent: &workflow_span, Level::INFO, "serializing").in_scope(|| 
        bincode::serialize(&grouped_results).unwrap()
    );
    let _: () = redis.set("grouped_results", &msg).instrument(span!(parent: &workflow_span, Level::INFO, "saving")).await.unwrap();
    let msg: Vec<u8> = redis.get("grouped_results").instrument(span!(parent: &workflow_span, Level::INFO, "loading")).await.unwrap();
    let grouped_results: Vec<Result> = span!(parent: &workflow_span, Level::INFO, "deserializing").in_scope(|| 
        bincode::deserialize(&msg).unwrap()
    );

    // sort 
    let span = span!(parent: &workflow_span, Level::INFO, "sorting");
    let _entered = span.enter();
    let mut sorted_results = grouped_results;
    sort_by_total_revenue(&mut sorted_results);
}

#[derive(Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
struct Request {
    mode: String,
    records: i32, 
    chunk_size: usize,
}

#[async_handler]
pub async fn handle(_stub: AgentStub, args: String) -> String {
    let req = serde_json::from_str::<Request>(&args).unwrap();
    let num_uvs = req.records;     // how many visits
    let num_ranks = num_uvs / 20; // how many websites
    let groups = num_uvs / 10;     // how many users
    let chunk_size = req.chunk_size;

    let (ranks, urls) = gen_rankings(num_ranks);
    let uvs = gen_user_visits(num_uvs, groups, &urls);

    let uvs_msg = bincode::serialize(&uvs).unwrap();
    let ranks_msg = bincode::serialize(&ranks).unwrap();
    let data_size = uvs_msg.len() + ranks_msg.len();

    let start_time = std::time::Instant::now();

    match req.mode.as_str() {
        "chitu" => chitu_running(uvs, ranks, chunk_size).await,
        "fusion" | "local" => chitu_running(uvs, ranks, num_ranks as usize).await,
        "redis" => redis_running(uvs, ranks).await, 
        _ => panic!("unknown mode"),
    }
    let duration = start_time.elapsed();
    format!("{}.{:03} seconds, {}MB data\n", duration.as_secs(), duration.subsec_millis(), data_size / (1024 * 1024))
}

