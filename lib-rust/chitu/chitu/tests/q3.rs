use chitu::diff::{Change, Diff};
use chrono::prelude::*;
use chrono::DateTime;
use chrono::Duration;
use chrono::Local;
use rand::Rng;
use serde::{Deserialize, Serialize};
use std::fmt::Debug;
use std::net::IpAddr;

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

#[derive(Debug, Clone, Deserialize, Serialize)]
struct ResultValue {
    total_revenue: f64,
    avg_rank: f64,
    occurence: i32,
}

#[derive(Debug, Clone, Deserialize, Serialize, PartialEq)]
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
    fn apply(self, v: &mut Self::Value) {
        v.extend(self.0);
    }
}

use std::collections::HashMap;
use std::hash::Hash;
use std::sync::Arc;

#[test]
#[ignore]
fn test_group_by() {
    let v = vec![1, 2, 1, 1, 2, 2, 4, 3, 5, 3, 5, 3, 3];
    let groups = group_into_map(v, |x| x % 2);
    println!("{:?}", groups);
}

fn group_into_map<T, K, F>(items: Vec<T>, f: F) -> HashMap<K, Vec<T>>
where
    F: Fn(&T) -> K,
    K: Eq + Hash,
{
    let mut groups = HashMap::new();

    for item in items {
        groups.entry(f(&item)).or_insert(Vec::new()).push(item);
    }

    groups
}

#[test]
#[ignore]
fn test_merge_map() {
    let mut m1 = HashMap::new();
    m1.insert("a", 1);
    m1.insert("b", 2);
    m1.insert("c", 3);

    let mut m2 = HashMap::new();
    m2.insert("a", 2);
    m2.insert("b", 3);
    m2.insert("d", 4);

    merge_map_or_insert(&mut m1, &m2, |v1, v2| v1 + v2, || 0);
    println!("{:?}", m1);
}

fn merge_map_or_insert<F, K, V>(
    m1: &mut HashMap<K, V>,
    m2: &HashMap<K, V>,
    f: F,
    default: impl Fn() -> V,
) where
    K: Hash + Eq + Clone,
    F: Fn(&V, &V) -> V,
{
    for (k, v) in m2 {
        m1.entry(k.clone())
            .and_modify(|v1| *v1 = f(v1, &v))
            .or_insert(f(&default(), v));
    }
}

#[test]
#[ignore]
fn test_gen_ip() {
    let ip = gen_random_ip();
    println!("{}", ip);
    assert!(ip.parse::<IpAddr>().is_ok());
}

fn gen_random_ip() -> String {
    let ip = vec![
        rand::thread_rng().gen_range(0..255).to_string(),
        rand::thread_rng().gen_range(0..255).to_string(),
        rand::thread_rng().gen_range(0..255).to_string(),
        rand::thread_rng().gen_range(0..255).to_string(),
    ];
    ip.join(".")
}

#[test]
#[ignore]
fn test_gen_url() {
    let url = gen_random_url();
    println!("{}", url);
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

#[test]
#[ignore]
fn test_gen_time() {
    let start_time: DateTime<Local> = Local.with_ymd_and_hms(1970, 1, 1, 0, 0, 0).unwrap();
    let end_time: DateTime<Local> = Local.with_ymd_and_hms(2020, 1, 1, 0, 0, 0).unwrap();
    let time = gen_random_time(start_time, end_time);
    println!("{}", time);
}

fn gen_random_time(start: DateTime<Local>, end: DateTime<Local>) -> DateTime<Local> {
    let duration = end - start;
    let secs = duration.num_seconds();
    let rand_secs = rand::thread_rng().gen_range(0..secs);
    start + Duration::seconds(rand_secs)
}

#[test]
#[ignore]
fn test_gen_uv() {
    let uv = gen_user_visits(
        10,
        2,
        &vec![
            "http://YQPuiNNlGS.com".to_string(),
            "http://FbmMxXIAcZ.com".to_string(),
        ],
    );
    println!("{:?}", uv);
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

#[test]
#[ignore]
fn test_gen_rank() {
    let (ranks, urls) = gen_rankings(10);
    println!("{:?}", ranks);
    println!("{:?}", urls);
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

#[test]
fn test_uv_size() {
    let (_ranks, urls) = gen_rankings(1);
    let uv = gen_user_visits(1, 1, &urls);
    let msg = bincode::serialize(&uv[0]).unwrap();
    println!("Len: {}", msg.len());
}

#[tokio::test(flavor = "multi_thread", worker_threads = 4)]
async fn compare_chitu_and_local() {
    let num_uvs = 100;  // how many visits
    let num_ranks = num_uvs / 20; // how many websites
    let groups = num_uvs / 10;     // how many users
    let chunk_size: usize = 100;

    let (ranks, urls) = gen_rankings(num_ranks);
    let uvs = gen_user_visits(num_uvs, groups, &urls);

    let res1 = chitu_running(uvs.clone(), ranks.clone(), chunk_size).await;
    let res2 = local_running(uvs.clone(), ranks.clone());

    assert_eq!(res1, res2);
}

async fn chitu_running(uvs: Vec<UserVisit>, ranks: Vec<Ranking>, chunk_size: usize) -> Vec<Result> {
    // init diffs
    let mut duvs = Diff::new();
    let mut filtered_uvs = Diff::new();
    let mut joined_results = Diff::new();

    let mut rank_map = HashMap::new();
    for rank in ranks {
        rank_map.insert(rank.url, rank.rank);
    }
    let rank_map = Arc::new(rank_map);

    // subscribe diffs
    let grouped_results = joined_results.onchange_state_async(
        HashMap::new(),
        |mut grouped_results: HashMap<String, ResultValue>, ch: VecExtend<JoinResult>| async move {
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
        },
    );

    let joined_results = filtered_uvs.onchange_state_async(
        joined_results,
        move |joined_results, ch: VecExtend<UserVisit>| {
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
            }
        },
    );

    let lower_time = Local.with_ymd_and_hms(1980, 1, 1, 0, 0, 0).unwrap().timestamp();
    let upper_time = Local.with_ymd_and_hms(2006, 1, 1, 0, 0, 0).unwrap().timestamp();

    let filtered_uvs = duvs.onchange_state_async(
        filtered_uvs,
        move |filtered_uvs, ch: VecExtend<UserVisit>| async move {
            let res = ch
                .0
                .into_iter()
                .filter(|uv| uv.time < upper_time && uv.time > lower_time)
                .collect::<Vec<_>>();
            filtered_uvs.change(VecExtend(res));
            filtered_uvs
        },
    );

    /*
    Following is the blocking version
    ```
    let filtered_uvs = duvs.onchange_blocking_state(
        filtered_uvs,
        move |filtered_uvs, ch: VecExtend<UserVisit>| {
            let res = ch
                .0
                .into_iter()
                .filter(|uv| uv.time < upper_time && uv.time > lower_time)
                .collect::<Vec<_>>();
            filtered_uvs.change(VecExtend(res));
            filtered_uvs
        },
    );
    ```
    */

    // let uvs_chunks: Vec<_> = uvs.chunks(chunk_size as usize).collect();
    // for uvs_chunk in uvs_chunks {
    //     duvs.change(VecExtend(uvs_chunk.to_vec()));
    // }

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

    println!(
        "Results 0: ip: {}, avg_rank: {}, total_revenue: {}",
        sorted_results[0].ip, sorted_results[0].avg_rank, sorted_results[0].total_revenue
    );

    sorted_results
}

#[tokio::test]
async fn test_redis() {
    let num_ranks = 10000;
    let num_uvs = 100000;
    let groups = 20000;
    let (ranks, urls) = gen_rankings(num_ranks);
    let uvs = gen_user_visits(num_uvs, groups, &urls);
    redis_running(uvs, ranks).await;
}

async fn redis_running(uvs: Vec<UserVisit>, ranks: Vec<Ranking>) {
    use redis::AsyncCommands;

    let client = redis::Client::open("redis://redis.openfaas:6379").unwrap();
    let mut redis = client.get_async_connection().await.unwrap();

    let mut rank_map = HashMap::new();
    for rank in ranks {
        rank_map.insert(rank.url, rank.rank);
    }

    let lower_time: DateTime<Local> = Local.with_ymd_and_hms(1980, 1, 1, 0, 0, 0).unwrap();
    let upper_time: DateTime<Local> = Local.with_ymd_and_hms(2006, 1, 1, 0, 0, 0).unwrap();

    // filter
    let filtered_uvs = uvs
        .into_iter()
        .filter(|uv| uv.time < upper_time.timestamp() && uv.time > lower_time.timestamp())
        .collect::<Vec<_>>();

    // trans
    let msg = serde_json::to_string(&filtered_uvs).unwrap();
    let _: () = redis.set("filtered_uvs", msg.as_bytes()).await.unwrap();
    let msg: Vec<u8> = redis.get("filtered_uvs").await.unwrap();
    let filtered_uvs: Vec<UserVisit> = serde_json::from_slice(&msg).unwrap();

    // join
    let joined_results = filtered_uvs
        .into_iter()
        .map(|uv| {
            let rank = rank_map.get(&uv.dest).unwrap();
            JoinResult {
                ip: uv.ip,
                revenue: uv.revenue,
                rank: *rank,
            }
        })
        .collect::<Vec<_>>();

    // trans
    let msg = serde_json::to_string(&joined_results).unwrap();
    let _: () = redis.set("joined_results", msg.as_bytes()).await.unwrap();
    let msg: Vec<u8> = redis.get("joined_results").await.unwrap();
    let joined_results: Vec<JoinResult> = serde_json::from_slice(&msg).unwrap();

    // group
    let grouped_results = joined_results
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
        .collect::<Vec<_>>();

    // trans
    let msg = serde_json::to_string(&grouped_results).unwrap();
    let _: () = redis.set("grouped_results", msg.as_bytes()).await.unwrap();
    let msg: Vec<u8> = redis.get("grouped_results").await.unwrap();
    let grouped_results: Vec<Result> = serde_json::from_slice(&msg).unwrap();

    // sort 
    let mut sorted_results = grouped_results;
    sort_by_total_revenue(&mut sorted_results);
}

fn local_running(uvs: Vec<UserVisit>, ranks: Vec<Ranking>) -> Vec<Result> {
    let mut rank_map = HashMap::new();
    for rank in ranks {
        rank_map.insert(rank.url, rank.rank);
    }

    let lower_time: DateTime<Local> = Local.with_ymd_and_hms(1980, 1, 1, 0, 0, 0).unwrap();
    let upper_time: DateTime<Local> = Local.with_ymd_and_hms(2006, 1, 1, 0, 0, 0).unwrap();

    // filter
    let filtered_uvs = uvs
        .into_iter()
        .filter(|uv| uv.time < upper_time.timestamp() && uv.time > lower_time.timestamp())
        .collect::<Vec<_>>();

    // join
    let joined_results = filtered_uvs
        .into_iter()
        .map(|uv| {
            let rank = rank_map.get(&uv.dest).unwrap();
            JoinResult {
                ip: uv.ip,
                revenue: uv.revenue,
                rank: *rank,
            }
        })
        .collect::<Vec<_>>();

    // group
    let grouped_results = joined_results
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
        .collect::<Vec<_>>();

    // sort 
    let mut sorted_results = grouped_results;
    sort_by_total_revenue(&mut sorted_results);

    sorted_results
}