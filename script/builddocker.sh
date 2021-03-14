#!/bin/bash

set -e

docker build -t backend:latest -f ./Dockerfile .