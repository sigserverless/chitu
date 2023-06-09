use serde::{Deserialize, Serialize};

use crate::utils::{call_json, cast_json};

pub async fn update_agent(ip: String, key: String, dag_id: String, msg: bytes::Bytes) {
    let url = format!("http://{}:8080/update", ip);
    let _res = reqwest::Client::new()
        .post(url)
        .query(&[("key", key), ("dagId", dag_id)])
        .body(msg)
        .send()
        .await
        .unwrap();
}

#[derive(Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct EndReq {
    pub key: String,
    pub dag_id: String,
}

pub async fn end_agent(ip: String, key: String, dag_id: String) {
    let url = format!("http://{}:8080/end", ip);
    let req = EndReq { key, dag_id };
    let res = cast_json(url.as_str(), req);
    res.await.unwrap()
}

const COORDINATOR_URL: &str = "http://coordinator.openfaas-fn:8080";

#[derive(Serialize)]
#[serde(rename_all = "camelCase")]
pub struct CoordPutReq {
    key: String,
    dag_id: String,
}

#[derive(Deserialize)]
pub struct CoordPutRes {
    pub ips: Vec<String>,
}

pub async fn coordinator_put(key: String, dag_id: String) -> CoordPutRes {
    let url = format!("{}/put", COORDINATOR_URL);
    let req = CoordPutReq { key, dag_id };
    let res = call_json::<CoordPutReq, CoordPutRes>(url.as_str(), req);
    res.await.unwrap()
}

#[derive(Serialize)]
#[serde(rename_all = "camelCase")]
pub struct CoordGetReq {
    key: String,
    dag_id: String,
}

pub async fn coordinator_get(key: String, dag_id: String) {
    let url = format!("{}/get", COORDINATOR_URL);
    let req = CoordGetReq { key, dag_id };
    let res = cast_json(url.as_str(), req);
    res.await.unwrap()
}

#[derive(Serialize)]
#[serde(rename_all = "camelCase")]
pub struct CoordTriggerReq {
    dag_id: String,
    fname: String,
    args: String,
}

pub async fn coordinator_trigger(dag_id: String, fname: String, args: String) {
    let url = format!("{}/trigger", COORDINATOR_URL);
    let req = CoordTriggerReq {
        dag_id,
        fname,
        args,
    };
    let res = cast_json(url.as_str(), req);
    res.await.unwrap()
}

pub async fn get_self_ip() -> Result<String, reqwest::Error> {
    let url = format!("{}/ip", COORDINATOR_URL);
    let res = reqwest::Client::new().post(url).send().await?;
    res.text().await
}
