##!/bin/bash
#
#set -e
##echo "\n" | aws configure
##echo -ne '\n\n\n\n' | aws configure
#cat << EOF > ~/.aws/config
#[default]
#region=us-east-2
#output=json
#EOF
#
#cat << EOF > ~/.aws/credentials
#[default]
#aws_access_key_id = dummy
#aws_secret_access_key = dummy
#EOF
#
#aws dynamodb --region us-east-2 --endpoint-url http://dynamodb:8000 create-table \
#  --cli-input-json file:///tables/profile.json
#
#aws dynamodb --region us-east-2 --endpoint-url http://dynamodb:8000 create-table \
#  --cli-input-json file:///tables/whitelist_profile.json
#
#aws dynamodb --region us-east-2 --endpoint-url http://dynamodb:8000 create-table \
#  --cli-input-json file:///tables/hive.json
#
#aws dynamodb --region us-east-2 --endpoint-url http://dynamodb:8000 create-table \
#  --cli-input-json file:///tables/hive_post.json
#
#aws dynamodb --region us-east-2 --endpoint-url http://dynamodb:8000 create-table \
#  --cli-input-json file:///tables/hive_post_reply.json
#
#  aws dynamodb --region us-east-2 --endpoint-url http://dynamodb:8000 create-table \
#  --cli-input-json file:///tables/post_comment_track.json
#
#aws dynamodb --region us-east-2 --endpoint-url http://dynamodb:8000 list-tables