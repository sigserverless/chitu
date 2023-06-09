use chitu_macros::async_handler;

#[async_handler]
pub async fn hello(args: String) -> String {
    args
}

#[test]
fn my_test() {
    let y = hello("hello".to_string());
    assert_eq!(y, "hello");
}