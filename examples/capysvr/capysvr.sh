docker run --rm \
    --name capyfile_server \
    --mount type=bind,source=./service-definition.json,target=/etc/capyfile/service-definition.json \
    --env CAPYFILE_SERVICE_DEFINITION_FILE=/etc/capyfile/service-definition.json \
    --env AWS_ENDPOINT=s3.amazonaws.com \
    --env AWS_REGION=us-west-1 \
    --secret aws_access_key_id \
    --secret aws_secret_access_key \
    -p 8024:8024 \
    capyfile/capysvr:latest
