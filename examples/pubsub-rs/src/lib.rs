use chitu::stub::AgentStub;
use chitu::diff::*;
use chitu_macros::async_handler;

#[async_handler]
pub async fn handle(stub: AgentStub, args: String) -> String {
    let mut export_obj: Diff<Vec<Vec<i32>>, VecChange<Vec<i32>>> = Diff::new();
    let mut import_obj:  Diff<Vec<Vec<i32>>, VecChange<Vec<i32>>> = Diff::new();

    stub.export("x", &mut export_obj).await;

    export_obj.onchange(move |ch| {
        match ch {
            VecChange::Push(_) => println!("Export Pushed"),
        }
    });

    import_obj.onchange(move |ch| {
        match ch {
            VecChange::Push(mut v) => { v.sort(); println!("Import: {}", v[0]) },
        }
    });

    let handle = stub.import("x", import_obj).await;

    let mut big_vector = vec![];
    for i in 0..1000000 {
        big_vector.push(65535 - i);
    } 

    for _i in 0..5 {
        export_obj.push(big_vector.clone());
    }

    export_obj.end().await;

    let import_obj = handle.await.unwrap();
    
    format!("Imported: {}, args: {}", import_obj.end().await.len(), args)
}