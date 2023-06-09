#[cfg(test)]
mod tests {
    use std::sync::{Arc, Mutex};

    use chitu::{agent::start_server, diff::*, stub::AgentStub, usercode::Usercode};

    #[tokio::test]
    async fn basic_use() {
        let mut dv = Diff::new();
        let len = Arc::new(Mutex::new(0));
        let len_clone = len.clone();

        dv.onchange(move |ch| {
            let mut len = len_clone.lock().unwrap();
            match ch {
                VecChange::Push(_) => *len += 1,
            }
        });

        dv.push(1).await;
        dv.push(2).await;
        dv.push(3).await;

        dv.end().await;

        let res = len.lock().unwrap();

        assert_eq!(*res, 3);
    }

    #[tokio::test]
    async fn get_state() {
        let mut dv = Diff::new();

        let len = dv.onchange_state(0, |state, ch| match ch {
            VecChange::Push(_) => state + 1,
        });

        dv.push(1).await;
        dv.push(2).await;
        dv.push(3).await;

        dv.end().await;

        let res = len.await.unwrap();

        assert_eq!(res, 3);
    }

    #[tokio::test]
    async fn async_without_state() {
        let mut dv = Diff::new();

        dv.onchange_async(|ch| async {
            match ch {
                VecChange::Push(_) => println!("Push"),
            }
        });

        dv.push(1).await;
        dv.push(2).await;
        dv.push(3).await;

        dv.end().await;

        assert_eq!(3, 3);
    }

    #[tokio::test]
    async fn async_with_state() {
        let mut dv = Diff::new();

        let len = dv.onchange_state_async(0, |state, ch| async move {
            match ch {
                VecChange::Push(_) => state + 1,
            }
        });

        dv.push(1).await;
        dv.push(2).await;
        dv.push(3).await;

        dv.end().await;
        let len = len.await.unwrap();

        assert_eq!(len, 3);
    }

    #[tokio::test]
    #[ignore]
    async fn max_payload_size() {
        start_server(
            Usercode::Native(handler),
            ("127.0.0.1", 3000),
            ("127.0.0.1", 3001),
        )
        .await;
    }

    fn handler(_stub: AgentStub, _args: String) -> String {
        "hello".to_string()
    }

    #[tokio::test]
    #[ignore]
    async fn max_payload_size_client() {
        let start_time = std::time::Instant::now();
        let msg = "a".repeat(1024 * 1024);
        let msg: bytes::Bytes = msg.into();
        let url = "localhost:31112/function/pubsub-rs";
        let res = reqwest::Client::new().post(url).body(msg).send().await;
        let duration = start_time.elapsed();
        println!("{}, duration: {:?}", res.unwrap().status(), duration);
    }

    #[tokio::test(flavor = "multi_thread", worker_threads = 3)]
    #[ignore]
    async fn async_onchange_compute() {
        let mut dv = Diff::new();
        let size = 400000000i128;

        dv.onchange(move |ch| {
            match ch {
                VecChange::Push(_) => {
                    let mut q = 16;
                    for i in 0..size {
                        q = q * i;
                        q = q / (i + 1);
                        // println!("print {}", i);
                    }
                }
            }
        });

        dv.onchange(move |ch| {
            match ch {
                VecChange::Push(_) => {
                    let mut q = 16;
                    for i in 0..size {
                        q = q * i;
                        q = q / (i + 1);
                        // println!("print {}", i);
                    }
                }
            }
        });

        dv.push(1).await;

        dv.end().await;

        assert_eq!(3, 3);
    }
}
