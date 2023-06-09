use chitu::diff::{Change, Diff};
use rand::prelude::*;
use serde::{Deserialize, Serialize};
use std::fmt::Debug;

#[derive(Deserialize, Serialize, Debug, Clone)]
struct VecExtend<T>(Vec<T>);

impl<T: Clone + Debug> Change for VecExtend<T> {
    type Value = Vec<T>;
    fn apply(self, v: &mut Self::Value) {
        v.extend(self.0);
    }
}

const SIZE: i32 = 10000;
const BATCH_SIZE: usize = 100;

#[tokio::test]
pub async fn test_inc_sort() {
    let mut rng = rand::thread_rng();
    let mut v: Vec<i32> = (0..SIZE).collect();

    v.shuffle(&mut rng);
    let vs: Vec<_> = v.chunks(BATCH_SIZE).collect();

    let mut dv = Diff::new();
    let sorted = dv.onchange_state(Vec::new(), |state, ch: VecExtend<i32>| {
        let s = merge_sort(ch.0);
        let res = merge(state, s);
        res
    });

    for v in vs {
        dv.change(VecExtend(v.to_vec()));
    }

    dv.end().await;
    let sorted = sorted.await.unwrap();

    assert_eq!(sorted.len() as i32, SIZE);
    assert!(is_sorted(&sorted));
}

fn is_sorted(vec: &[i32]) -> bool {
    for i in 1..vec.len() {
        if vec[i] < vec[i - 1] {
            return false;
        }
    }
    true
}

fn merge(v1: Vec<i32>, v2: Vec<i32>) -> Vec<i32> {
    let mut res = Vec::new();
    let mut i = 0;
    let mut j = 0;
    while i < v1.len() && j < v2.len() {
        if v1[i] < v2[j] {
            res.push(v1[i]);
            i += 1;
        } else {
            res.push(v2[j]);
            j += 1;
        }
    }
    while i < v1.len() {
        res.push(v1[i]);
        i += 1;
    }
    while j < v2.len() {
        res.push(v2[j]);
        j += 1;
    }
    res
}

fn merge_sort(v: Vec<i32>) -> Vec<i32> {
    if v.len() <= 1 {
        return v;
    }
    let mid = v.len() / 2;
    let left = merge_sort(v[..mid].to_vec());
    let right = merge_sort(v[mid..].to_vec());
    merge(left, right)
}
