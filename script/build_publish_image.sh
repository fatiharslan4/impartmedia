#!/bin/bash

set -e

aws_account_id=518740895671
aws_ecr_region=us-east-2
repository=impartwealth/backend
sha=$(git rev-parse --short HEAD)
image_name="${aws_account_id}.dkr.ecr.${aws_ecr_region}.amazonaws.com/${repository}:${sha}"

echo "building sha ${sha} and pushing as ${image_name}"

aws ecr get-login-password --region ${aws_ecr_region} | docker login --username AWS --password-stdin ${aws_account_id}.dkr.ecr.${aws_ecr_region}.amazonaws.com

docker build -t "${image_name}" -f ./Dockerfile .

docker push "${image_name}"

echo "done building sha ${sha} as, published as ${image_name}"