use serde::{de::DeserializeOwned, Serialize};

pub async fn call_json<Req, Res>(url: &str, req: Req) -> Result<Res, reqwest::Error>
where
    Req: Serialize,
    Res: DeserializeOwned,
{
    let res = reqwest::Client::new()
        .post(url)
        .json(&req)
        .send()
        .await?
        .json()
        .await?;
    Ok(res)
}

pub async fn cast_json<Req>(url: &str, req: Req) -> Result<(), reqwest::Error>
where
    Req: Serialize,
{
    let _res = reqwest::Client::new().post(url).json(&req).send().await?;
    Ok(())
}

pub async fn cast_bytes(url: &str, req: bytes::Bytes) -> Result<(), reqwest::Error> {
    let _res = reqwest::Client::new().post(url).body(req).send().await?;
    Ok(())
}
