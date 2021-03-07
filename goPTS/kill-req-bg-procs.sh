#!/bin/bash
set -x
sudo kill -9 $(ps axu | grep "request-generator" | awk '{print $2}' | xargs)
set +x
