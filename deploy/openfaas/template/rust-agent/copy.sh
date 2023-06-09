#!/bin/bash 

rsync -av --exclude 'target' --exclude 'Cargo.lock' $CHITU_HOME/lib-rust/chitu ./
cp -R $CHITU_HOME/deploy/certs/ .