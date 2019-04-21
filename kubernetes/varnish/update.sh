#!/usr/bin/env bash

set -ex

docker build . -t dmarby/picsum-photos-varnish:latest
docker push dmarby/picsum-photos-varnish:latest
