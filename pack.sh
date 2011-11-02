#!/bin/bash

set -e

rm -rf SemkiBot.zip
cur=`pwd`
tmp=`tempfile`
rm $tmp
cp -rf src $tmp
cd $tmp
rm *_test.go
zip -r $cur/SemkiBot.zip *

echo "Done. SemkiBot.zip has been created."