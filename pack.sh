#!/bin/bash

set -e

rm -rf SemkiBot.zip
cd src
zip -r ../SemkiBot.zip *

echo "Done. SemkiBot.zip has been created."