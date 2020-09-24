#!/usr/bin/env bash
set -e

DIR="$(cd "$(dirname "$0")" && pwd)"

url=$1

echo "Downloading $url"

mkdir -p "data/$(dirname $url)"

wget "$url" -O "data/$url"
