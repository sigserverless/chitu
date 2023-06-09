#[cfg(test)]
mod archive {
    use std::path::Path;
    use std::process::Stdio;
    use std::time::Duration;

    use actix_web::Responder;
    use nix::sys::stat::Mode;
    use nix::unistd::mkfifo;
    use serde::{Deserialize, Serialize};
    use tokio::fs::{self, OpenOptions};
    use tokio::io::{self, AsyncBufReadExt, AsyncWriteExt};
    use tokio::process::Command;
    use tokio::sync::mpsc;
    use tokio::time::sleep;

    #[derive(Deserialize, Serialize)]
    #[serde(rename_all = "camelCase")]
    #[serde(tag = "kind")]
    enum CommandHelper {
        Import {
            key: String,
            dag_id: String,
        },
        Export {
            key: String,
            dag_id: String,
        },
        #[serde(rename = "pub")]
        PubHelper {
            key: String,
            dag_id: String,
            len: i64,
        },
    }

    #[test]
    fn test_serde() {
        let command = CommandHelper::PubHelper {
            key: "k".to_string(),
            dag_id: "d".to_string(),
            len: 65,
        };
        let command_str = serde_json::to_string(&command).unwrap();
        println!("{}", command_str);

        let command = CommandHelper::Import {
            key: "k2".to_string(),
            dag_id: "d2".to_string(),
        };
        let command_str = serde_json::to_string(&command).unwrap();
        println!("{}", command_str);
    }

    #[tokio::test]
    #[ignore]
    async fn test_command() -> io::Result<()> {
        let handle = tokio::spawn(run_command_with_input(
            "/usr/bin/python3",
            &["./tests/handle.py"],
            "Tom",
        ));
        for x in 0..3 {
            println!("x = {}", x);
            sleep(Duration::from_secs(1)).await;
        }
        handle.await??;
        Ok(())
    }

    async fn run_command_with_input(cmd: &str, args: &[&str], input: &str) -> std::io::Result<()> {
        let mut child = Command::new(cmd)
            .args(args)
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .spawn()?;

        if let Some(mut stdin) = child.stdin.take() {
            stdin.write_all(input.as_bytes()).await?;
        }

        while let Some(output) = child.stdout.take() {
            let mut reader = tokio::io::BufReader::new(output);
            let mut buffer = String::new();
            while reader.read_line(&mut buffer).await? > 0 {
                println!("[sub] {}", buffer);
                buffer.clear();
            }
        }

        println!("Sub process ending");

        Ok(())
    }

    #[tokio::test]
    #[ignore]
    async fn test_fifo() -> io::Result<()> {
        testing_fifo().await?;
        Ok(())
    }

    async fn testing_fifo() -> io::Result<()> {
        let fifo_path = "/tmp/fifo";

        if !Path::new(fifo_path).exists() {
            mkfifo(fifo_path, Mode::S_IRWXU)?;
        }

        let write_handle = tokio::spawn(async move {
            let mut file = OpenOptions::new()
                .write(true)
                .open(fifo_path)
                .await
                .unwrap();
            for _i in 0..5 {
                file.write_all(b"Hello, world!\n").await.unwrap();
                file.write_all(b"\n").await.unwrap();
                sleep(Duration::from_secs(1)).await;
            }
        });

        let read_handle = tokio::spawn(async move {
            let file = OpenOptions::new().read(true).open(fifo_path).await.unwrap();
            let mut buffer = String::new();
            let mut reader = tokio::io::BufReader::new(file);
            while reader.read_line(&mut buffer).await.unwrap() > 0 {
                println!("Read: {}, len: {}", buffer.trim_end(), buffer.len());
                buffer.clear();
            }
            println!("Ending");
        });

        // sleep(Duration::from_secs(5)).await;

        write_handle.await?;
        read_handle.await?;

        fs::remove_file(fifo_path).await?;

        Ok(())
    }

    #[tokio::test]
    async fn test_async_fn_type() {
        let h = hello;
        consume_async_fn(h).await
    }

    async fn consume_async_fn<F, R>(f: F)
    where
        F: Fn() -> R,
        R: std::future::Future<Output = String>,
    {
        let s = f();
        println!("{}", s.await);
    }

    async fn hello() -> String {
        "hello".to_string()
    }

    #[tokio::test(flavor = "multi_thread", worker_threads = 2)]
    async fn test_optional_text_server() {
        actix_web::HttpServer::new(move || actix_web::App::new().service(index))
            .bind(("127.0.0.1", 3000))
            .unwrap()
            .run()
            .await
            .unwrap();
    }

    #[actix_web::post("/")]
    async fn index(text: String) -> impl Responder {
        std::thread::spawn(move ||
            tokio::runtime::Runtime::new()
                .unwrap()
                .block_on(workers())
        ).join().unwrap();
        format!("Hello {}!", text)
    }

    #[test]
    fn merge_hashmap() {
        use std::collections::HashMap;

        let mut map1 = HashMap::new();
        map1.insert("a", 1);
        map1.insert("b", 2);

        let mut map2 = HashMap::new();
        map2.insert("b", 3);
        map2.insert("c", 4);

        map1.extend(map2);
        println!("{:?}", map1);
    }

    #[derive(Serialize, Deserialize)]
    #[serde(rename_all = "camelCase")]
    struct Request {
        mode: String,
        num_ranks: i32,
        num_uvs: i32,
        groups: i32,
        chunk_size: usize,
    }

    #[test]
    #[ignore]
    fn test_deserialization() {
        let input = "
        {
            \"mode\": \"local\",
            \"numRanks\": 20,
            \"numUvs\": 20,
            \"groups\": 3,
            \"chunkSize\": 1
        }
        ";
        let req = serde_json::from_str::<Request>(input).unwrap();
        println!("{}", req.mode);
    }

    #[tokio::test(flavor = "multi_thread", worker_threads = 3)]
    #[ignore]
    async fn test_multicores() {
        workers().await;
    }

    async fn workers() {
        let mut handles = vec![];
        let size = 400000000i128;
        // let size = 10;
        for _ in 0..4 {
            let h = tokio::spawn(async move {
                sleep(Duration::from_secs(1)).await;
                let mut q = 16;
                for i in 0..size {
                    q = q * i;
                    q = q / (i + 1);
                    // println!("print {}", i);
                }
            });
            handles.push(h);
        }
        for h in handles {
            h.await.unwrap();
        }
    }

    #[tokio::test(flavor = "multi_thread", worker_threads = 4)]
    #[ignore]
    async fn test_runtime_workers() {
        std::thread::spawn(move ||
            tokio::runtime::Runtime::new()
                .unwrap()
                .block_on(workers())
        ).join().unwrap();
    }

    #[tokio::test]
    #[ignore]
    async fn test_receivers() {
        let (tx, mut rx) = mpsc::channel(1);
        tx.send(12).await.unwrap();
        // tx.send(13).await.unwrap();
        drop(tx);

        while let Some(v) = rx.recv().await {
            println!("received {}", v);
        }
        println!("end");
    }
}
