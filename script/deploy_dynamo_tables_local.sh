#!/bin/bash

set -e
#echo "\n" | aws configure
#echo -ne '\n\n\n\n' | aws configure
mkdir -p ~/.aws
cat << EOF > ~/.aws/config
[local]
region=us-east-2
output=json
EOF

cat << EOF > ~/.aws/credentials
[local]
aws_access_key_id = dummy
aws_secret_access_key = dummy
EOF

aws --profile local dynamodb --region us-east-2 --endpoint-url http://dynamodb:8000 create-table \
  --cli-input-json file:///tables/profile.json

aws --profile local dynamodb --region us-east-2 --endpoint-url http://dynamodb:8000 create-table \
  --cli-input-json file:///tables/whitelist_profile.json

aws --profile local dynamodb --region us-east-2 --endpoint-url http://dynamodb:8000 create-table \
  --cli-input-json file:///tables/hive.json

aws --profile local dynamodb --region us-east-2 --endpoint-url http://dynamodb:8000 create-table \
  --cli-input-json file:///tables/hive_post.json

aws --profile local dynamodb --region us-east-2 --endpoint-url http://dynamodb:8000 create-table \
  --cli-input-json file:///tables/hive_post_reply.json

aws --profile local dynamodb --region us-east-2 --endpoint-url http://dynamodb:8000 create-table \
  --cli-input-json file:///tables/post_comment_track.json

cat << EOF > allow.json
{
  "email": {
    "S": "dev.poster@test.com"
  },
  "impartWealthId": {
    "S": "1GuRwMnzwwRxE0phxifmPMHH9hX"
  },
  "screenName": {
    "S": "\"newscreenname\""
  }
}
EOF

aws --profile local dynamodb --region us-east-2 --endpoint-url http://dynamodb:8000 put-item --table-name local_whitelist_profile --item file://./allow.json
aws --profile local dynamodb --region us-east-2 --endpoint-url http://dynamodb:8000 list-tables