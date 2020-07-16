#!/bin/bash
# Testing framework: https://github.com/pgrange/bash_unit

setup_suite() {
  mkdir -p output

  docker network create test >/dev/null

  # start webgoat
  tar -xzvf fixtures/webgoat-data.tar.gz -C ./ >/dev/null 2>&1

  docker run --rm \
    -v "${PWD}/.webgoat-8.0.0.M21":/home/webgoat/.webgoat-8.0.0.M21 \
    --name goat \
    --network test \
    -d \
    registry.gitlab.com/gitlab-org/security-products/dast/webgoat-8.0@sha256:bc09fe2e0721dfaeee79364115aeedf2174cce0947b9ae5fe7c33312ee019a4e >/dev/null

  # start nginx container to use as proxy for domain validation
  docker run --rm \
    -v "${PWD}/fixtures/domain-validation/nginx.conf/nginx.conf":/etc/nginx/conf.d/default.conf \
    --name vulnerabletestserver \
    --network test \
    -d \
    nginx:1.17.6-alpine >/dev/null
}

teardown_suite() {
  docker rm -f goat vulnerabletestserver >/dev/null 2>&1
  docker network rm test >/dev/null 2>&1
  rm -r .webgoat-8.0.0.M21 >/dev/null
  true
}

test_webgoat() {
  docker run --rm -it \
    -v "${PWD}"/output:/browserker/output \
    --network test \
    "browserker" ./browserker crawl --config ./configs/webgoat.toml --dot output/webgoat.dot --profile *> debug.log
}
