docker run --rm \
    --name capyfile_cmd \
    --mount type=bind,source=./service-definition.json,target=/etc/capyfile/service-definition.json \
    --mount type=bind,source=./images,target=/app/images \
    --env CAPYFILE_SERVICE_DEFINITION_FILE=/etc/capyfile/service-definition.json \
    --env INPUT_READ_TARGET="/app/images/*" \
    --env INPUT_WRITE_DESTINATION="/app/images/converted" \
    capyfile/capycmd:latest images:convert
