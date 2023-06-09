extern crate proc_macro;

use proc_macro::TokenStream;
use quote::quote;
use syn::{parse_macro_input, ItemFn, Signature};

#[proc_macro_attribute]
pub fn async_handler(_attr: TokenStream, item: TokenStream) -> TokenStream {
    let input = parse_macro_input!(item as ItemFn);
    let ItemFn { vis, .. } = input;
    let Signature {
        ident,
        inputs,
        output,
        ..
    } = input.sig;

    let block = input.block;

    let output = quote! {
        #vis fn #ident (#inputs) #output {
            std::thread::spawn(move ||
                tokio::runtime::Runtime::new()
                    .unwrap()
                    .block_on(async #block)
            ).join().unwrap()
        }
    };

    TokenStream::from(output)
}
