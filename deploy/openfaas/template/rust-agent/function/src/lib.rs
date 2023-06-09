use chitu::stub::AgentStub;

pub fn handle(stub: AgentStub, args: String) -> String {
    args
}

// Or, if you want async handler: 
// 
// #[async_handler]
// pub async fn handle(stub: AgentStub, args: String) -> String {
//     args
// }