#!/bin/bash 


rsync -av --exclude 'target' --exclude 'Cargo.lock' $CHITU_HOME/lib-rust/chitu ./
rm -r ./lib-python
cp -R $CHITU_HOME/lib-python . 
cp -R $CHITU_HOME/deploy/certs/ .