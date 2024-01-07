# Examples

The below you can see a couple of examples of the service definition file along with the
commands to run the file processing pipeline.

If you want to see more detailed examples, please check these articles:
* [Do whatever you want with your files, and do it quickly](https://dev.to/sky003/do-whatever-you-want-with-your-files-and-do-it-quickly-4od2)
* [Integrate any command into your file-processing pipeline](https://dev.to/sky003/integrate-any-command-into-your-file-processing-pipeline-3jbh)

### Log archiver example

First example is a typical service definition file for the `capycmd` command line application.
This service definition setting up a pipeline that reads the log files from the filesystem,
uploads the files that are older than 30 days to S3-compatible storage and removes them.
```yaml
---
version: '1.2'
name: logs
processors:
  - name: archive
    operations:
      - name: filesystem_input_read
        params:
          target:
            sourceType: value
            source: "/var/log/rotated-logs/*"
      - name: file_time_validate
        params:
          maxMtime:
            sourceType: env_var
            source: MAX_LOG_FILE_AGE_RFC3339
      - name: s3_upload
        targetFiles: without_errors
        params:
          accessKeyId:
            sourceType: secret
            source: aws_access_key_id
          secretAccessKey:
            sourceType: secret
            source: aws_secret_access_key
          endpoint:
            sourceType: value
            source: s3.amazonaws.com
          region:
            sourceType: value
            source: us-east-1
          bucket:
            sourceType: env_var
            source: AWS_LOGS_BUCKET
      - name: filesystem_input_remove
        targetFiles: without_errors
```

Now when you have a service definition file, you can run the file processing pipeline.
```bash
# Provide service definition stored in the local filesystem 
# via CAPYFILE_SERVICE_DEFINITION_FILE=/etc/capyfile/service-definition.json
docker run \
    --name capyfile_cmd \
    --mount type=bind,source=./service-definition.yml,target=/etc/capyfile/service-definition.yml \
    --mount type=bind,source=/var/log/rotated-logs,target=/var/log/rotated-logs \
    --env CAPYFILE_SERVICE_DEFINITION_FILE=/etc/capyfile/service-definition.yml \
    --env MAX_LOG_FILE_AGE_RFC3339=$(date -d "30 days ago" -u +"%Y-%m-%dT%H:%M:%SZ") \
    --env AWS_LOGS_BUCKET=logs \
    --secret aws_access_key_id \
    --secret aws_secret_access_key \
    capyfile/capycmd:latest --concurrency logs:archive
```

The command produces the output that can look like this (the output has weird order because
it is a result of concurrent processing):
```
Running logs:archive service processor...

[/var/log/rotated-logs/access-2023-08-27.log] filesystem_input_read FINISHED file read finished
[/var/log/rotated-logs/access-2023-08-28.log] filesystem_input_read FINISHED file read finished
[/var/log/rotated-logs/access-2023-09-27.log] filesystem_input_read FINISHED file read finished
[/var/log/rotated-logs/access-2023-09-28.log] filesystem_input_read FINISHED file read finished
[/var/log/rotated-logs/access-2023-09-29.log] filesystem_input_read FINISHED file read finished
[/var/log/rotated-logs/access-2023-08-28.log] file_time_validate STARTED file time validation started
[/var/log/rotated-logs/access-2023-08-28.log] file_time_validate FINISHED file time is valid
[/var/log/rotated-logs/access-2023-08-27.log] file_time_validate STARTED file time validation started
[/var/log/rotated-logs/access-2023-09-27.log] file_time_validate STARTED file time validation started
[/var/log/rotated-logs/access-2023-08-27.log] file_time_validate FINISHED file time is valid
[/var/log/rotated-logs/access-2023-09-27.log] file_time_validate FINISHED file mtime is too new
[/var/log/rotated-logs/access-2023-09-29.log] file_time_validate STARTED file time validation started
[/var/log/rotated-logs/access-2023-09-27.log] s3_upload SKIPPED skipped due to "without_errors" target files policy
[/var/log/rotated-logs/access-2023-09-28.log] file_time_validate STARTED file time validation started
[/var/log/rotated-logs/access-2023-09-29.log] file_time_validate FINISHED file mtime is too new
[/var/log/rotated-logs/access-2023-08-28.log] s3_upload STARTED S3 file upload has started
[/var/log/rotated-logs/access-2023-09-28.log] file_time_validate FINISHED file mtime is too new
[/var/log/rotated-logs/access-2023-09-29.log] s3_upload SKIPPED skipped due to "without_errors" target files policy
[/var/log/rotated-logs/access-2023-08-27.log] s3_upload STARTED S3 file upload has started
[/var/log/rotated-logs/access-2023-09-27.log] filesystem_input_remove SKIPPED skipped due to "without_errors" target files policy
[/var/log/rotated-logs/access-2023-09-29.log] filesystem_input_remove SKIPPED skipped due to "without_errors" target files policy
[/var/log/rotated-logs/access-2023-09-28.log] s3_upload SKIPPED skipped due to "without_errors" target files policy
[/var/log/rotated-logs/access-2023-09-28.log] filesystem_input_remove SKIPPED skipped due to "without_errors" target files policy
[/var/log/rotated-logs/access-2023-08-27.log] s3_upload FINISHED S3 file upload has finished
[/var/log/rotated-logs/access-2023-08-28.log] s3_upload FINISHED S3 file upload has finished
[/var/log/rotated-logs/access-2023-08-27.log] filesystem_input_remove STARTED file remove started
[/var/log/rotated-logs/access-2023-08-27.log] filesystem_input_remove FINISHED file remove finished
[/var/log/rotated-logs/access-2023-08-28.log] filesystem_input_remove STARTED file remove started
[/var/log/rotated-logs/access-2023-08-28.log] filesystem_input_remove FINISHED file remove finished
...
````

### Document uploader example

This service definition setting up a pipeline that allows .pdf, .doc and .docx files that are
less than 10MB. Valid files will be uploaded to S3-compatible storage.
```yaml
---
version: '1.2'
name: documents
processors:
  - name: upload
    operations:
      - name: http_multipart_form_input_read
      - name: file_size_validate
        params:
          maxFileSize:
            sourceType: value
            source: 1048576
      - name: file_type_validate
        params:
          allowedMimeTypes:
            sourceType: value
            source:
              - application/pdf
              - application/msword
              - application/vnd.openxmlformats-officedocument.wordprocessingml.document
      - name: s3_upload
        params:
          accessKeyId:
            sourceType: secret
            source: aws_access_key_id
          secretAccessKey:
            sourceType: secret
            source: aws_secret_access_key
          endpoint:
            sourceType: etcd
            source: "/services/upload/aws_endpoint"
          region:
            sourceType: etcd
            source: "/services/upload/aws_region"
          bucket:
            sourceType: env_var
            source: AWS_DOCUMENTS_BUCKET

```

Now when you have a service definition file, you can run the file processing pipeline
available over HTTP.
```bash
# Provide service definition stored in the local filesystem 
# via CAPYFILE_SERVICE_DEFINITION_FILE=/etc/capyfile/service-definition.yml
docker run \
    --name capyfile_server \
    --mount type=bind,source=./service-definition.yml,target=/etc/capyfile/service-definition.yml \
    --env CAPYFILE_SERVICE_DEFINITION_FILE=/etc/capyfile/service-definition.yml \
    --env AWS_DOCUMENTS_BUCKET=documents \
    --secret aws_access_key_id \
    --secret aws_secret_access_key \
    -p 8024:80 \
    capyfile/capysvr:latest

# Or you can provide the service definition stored in the remote host 
# via CAPYFILE_SERVICE_DEFINITION_URL=https://example.com/service-definition.json
docker run \
    --name capyfile_server \
    --env CAPYFILE_SERVICE_DEFINITION_URL=https://example.com/service-definition.json \
    --env AWS_DOCUMENTS_BUCKET=documents \
    --secret aws_access_key_id \
    --secret aws_secret_access_key \
    -p 8024:80 \
    capyfile/capysvr:latest
```

If you want to load parameters from etcd, you can provide the etcd connection parameters via
environment variables:
```
ETCD_ENDPOINTS=["etcd1:2379","etcd2:22379","etcd3:32379"]
ETCD_USERNAME=etcd_user
ETCD_PASSWORD=etcd_password
```

Now it is ready to accept and process the files.
```bash
# upload and process single file
curl -F "file1=@$HOME/Documents/document.pdf" http://localhost/upload/document 

# upload and process request body
curl --data-binary "@$HOME/Documents/document.pdf" http://localhost/upload/document 

# upload and process multiple files
curl -F "file1=@$HOME/Documents/document.pdf" http://localhost/upload/document 
curl \
    -F "file1=@$HOME/Documents/document.pdf" \
    -F "file3=@$HOME/Documents/document.docx" \
    -F "file3=@$HOME/Documents/very-big-document.pdf" \
    -F "file4=@$HOME/Documents/program.run" \
    http://localhost/upload/document 
```

The service returns json response of this format (example for multiple files upload above):
```json
{
  "status": "PARTIAL",
  "code": "PARTIAL",
  "message": "successfully uploaded 2 of 4 files",
  "files": [
    {
      "url": "https://documents.storage.example.com/documents/abcdKDNJW_DDWse.pdf",
      "filename": "abcdKDNJW_DDWse.pdf",
      "originalFilename": "document.pdf",
      "mime": "application/pdf",
      "size": 5892728,
      "status": "SUCCESS",
      "code": "FILE_SUCCESSFULLY_UPLOADED",
      "message": "file successfully uploaded"
    },
    {
      "url": "https://documents.storage.example.com/documents/abcdKDNJW_DDWsd.docx",
      "filename": "abcdKDNJW_DDWsd.docx",
      "originalFilename": "document.docx",
      "mime": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
      "size": 3145728,
      "status": "SUCCESS",
      "code": "FILE_SUCCESSFULLY_UPLOADED",
      "message": "file successfully uploaded"
    }
  ],
  "errors": [
    {
      "originalFilename": "very-big-document.pdf",
      "status": "ERROR",
      "code": "FILE_IS_TOO_BIG",
      "message": "file size can not be greater than 10 MB"
    },
    {
      "originalFilename": "program.run",
      "status": "ERROR",
      "code": "FILE_MIME_TYPE_IS_NOT_ALLOWED",
      "message": "file MIME type \"application/x-makeself\" is not allowed"
    }
  ],
  "meta": {
    "totalUploads": 4,
    "successfulUploads": 2,
    "failedUploads": 2
  }
}
```