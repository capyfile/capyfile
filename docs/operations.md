# Operations

Operation represents a single action that can be performed on the input files. It
can be considered as a single step in the file-processing pipeline. 

The operations can be placed in any order that makes sense for you.

The list of available operations:
* [http_multipart_form_input_read](#http_multipart_form_input_read) - read the files from the HTTP request body as `multipart/form-data`
* [http_octet_stream_input_read](#http_octet_stream_input_read) - read the files from the HTTP request body as `application/octet-stream`
* [filesystem_input_read](#filesystem_input_read) - read the files from the filesystem
* [filesystem_input_write](#filesystem_input_write) - write the files to the filesystem
* [filesystem_input_remove](#filesystem_input_remove) - remove the files from the filesystem
* [input_forget](#input_forget) - forget the files
* [file_size_validate](#file_size_validate) - check file size
* [file_type_validate](#file_type_validate) - check file MIME type
* [file_time_validate](#file_time_validate) - check file time stat
* [exiftool_metadata_cleanup](#exiftool_metadata_cleanup) - clear file metadata if possible (require exiftool)
* [image_convert](#image_convert) - convert image to another format (require libvips)
* [s3_upload](#s3_upload) - upload file to S3-compatible storage
* [command_exec](#command_exec) - execute arbitrary command

## Operation parameters

Every operation supports the same root level parameters.

| Name            | Type    | Description                                                                                                                                        |
|-----------------|---------|----------------------------------------------------------------------------------------------------------------------------------------------------|
| `name`          | string  | Operation name.                                                                                                                                    |
| `params`        | ?object | Operation parameters that are specific to this exact operation.                                                                                    |
| `targetFiles`   | string  | What files the operation can process. <br/>Possible values: `without_errors` (default), `with_errors`, `all`.                                      |
| `cleanupPolicy` | string  | What to do with the files created by the operation when it's time to do the cleanup. <br/>Possible values: `keep_files` (default), `remove_files`. |
| `maxPacketSize` | int     | Maximum size of the operation's input, which is the number of files the operation can process at once (default: 0 - unlimited).                    |

## Operation parameters sources

The parameters that are specific to the operation can be sourced from different places.

| Source type   | Source                     | Description                                                                                                   |
|---------------|----------------------------|---------------------------------------------------------------------------------------------------------------|
| `value`       | Any value.                 | Retrieve parameter value from the pipeline configuration file.<br/> Source example: `/home/user/Videos/*.mp4` |
| `env_var`     | Environment variable name. | Retrieve parameter value from the environment variable.<br/> Source example: `AWS_REGION`                     |
| `file`        | File path.                 | Retrieve parameter value from the file.<br/> Source example: `/home/user/.capyfile/default-output-dir`        |
| `secret`      | Secret name.               | Retrieve parameter value from the secret.<br/> Source example: `aws_access_key_id`                            |
| `http_get`    | GET parameter name.        | Retrieve parameter value from the HTTP GET parameter.<br/> Source example: `aws_bucket`                       |
| `http_post`   | POST parameter name.       | Retrieve parameter value from the HTTP POST parameter.<br/> Source example: `capyfile_aws_bucket`             |
| `http_header` | Header name.               | Retrieve parameter value from the HTTP header.<br/> Source example: `X-Capyfile-Aws-Bucket`                   |
| `etcd`        | Key path.                  | Retrieve parameter value from the etcd key.<br/> Source example: `/capyfile/aws/bucket`                       |

### Example

The following example shows operation root parameters in action.

```yaml
operations:
  - name: filesystem_input_read
    # we don't want to remove the original files (default behavior)
    cleanupPolicy: keep_files
    params:
      target:
        sourceType: value
        source: "/home/user/Pictures/*"
  - name: image_convert
    # remove all converted files after they are no longer needed
    cleanupPolicy: remove_files
    params:
      toMimeType: 
        sourceType: value
        source: image/jpeg
      quality: 
        sourceType: value
        source: high
  - name: s3_upload
    # upload only files without errors (default behavior)
    targetFiles: without_errors
    # upload 10 files at once in the parallel
    maxPacketSize: 10
    params:
      accessKeyId: 
        sourceType: secret
        source: aws_access_key_id
      secretAccessKey: 
        sourceType: secret
        source: aws_secret_access_key
      region: 
        sourceType: env_var
        source: AWS_REGION
      bucket: 
        sourceType: value
        source: my-bucket
      endpoint: 
        sourceType: env_var
        source: AWS_ENDPOINT
```

## Available operations

### http_multipart_form_input_read

Read the files from the HTTP request body as `multipart/form-data`. Available for `capysvr` only.

### http_octet_stream_input_read

Read the files from the HTTP request body as `application/octet-stream`. Available for `capysvr` only.

### filesystem_input_read

Read the files from the local filesystem.

#### Parameters

| Name     | Type   | Description                                          |
|----------|--------|------------------------------------------------------|
| `target` | string | Path to the file or directory. Support glob pattern. |

#### Example

```yaml
name: filesystem_input_read
params:
  target: 
    sourceType: value
    source: /home/user/Videos/*.mp4
```

### filesystem_input_write

Write the files to the local filesystem.

#### Parameters

| Name                  | Type   | Description                                                         |
|-----------------------|--------|---------------------------------------------------------------------|
| `destination`         | string | Path to the directory to write the files.                           |
| `useOriginalFilename` | bool   | Whether to use the original filename or the generated one (nanoid). |

#### Example

```yaml
name: filesystem_input_write
params:
  destination: 
    sourceType: value
    source: /home/user/Videos/converted
  useOriginalFilename: 
    sourceType: value
    source: true
```

### filesystem_input_remove

Remove the files from the local filesystem.

#### Parameters

| Name                 | Type | Description                                                                |
|----------------------|------|----------------------------------------------------------------------------|
| `removeOriginalFile` | bool | Whether to remove the original file along with the processed one (if any). |

#### Example

```yaml
name: filesystem_input_remove
params:
  removeOriginalFile: 
    sourceType: value
    source: true
```

### input_forget

Forget the files.

Can be used when the current input should not be passed to the next operation.
Combined with the file target policy, it can be used to forget certain input,
such as the input with errors or without. Keep in mind that in such case, the
cleanup policy won't be applied because capyfile stops tracking these files.

#### Example

```yaml
name: input_forget
```

### file_size_validate

Check file size.

If the file size is not valid, capyfile attaches the error to the processable file.

#### Parameters

| Name          | Type   | Description                 |
|---------------|--------|-----------------------------|
| `minFileSize` | number | Minimum file size in bytes. |
| `maxFileSize` | number | Maximum file size in bytes. |

#### Example

```yaml
name: file_size_validate
params:
  minFileSize: 
    sourceType: value
    source: 1000000
  maxFileSize: 
    sourceType: value
    source: 100000000
```

### file_type_validate

Check file MIME type.

If the file MIME type is not valid, capyfile attaches the error to the processable file.

#### Parameters

| Name               | Type     | Description                 |
|--------------------|----------|-----------------------------|
| `allowedMimeTypes` | string[] | List of allowed MIME types. |

#### Example

```yaml
name: file_type_validate
params:
  allowedMimeTypes: 
    sourceType: value
    source: ["image/jpeg", "image/png"]
```

### file_time_validate

Check file time stat.

If the file time stat is not valid, capyfile attaches the error to the processable file.

#### Parameters

| Name          | Type | Description                                  |
|---------------|------|----------------------------------------------|
| `minAtime`    | int  | Minimum atime that can be parsed as RFC3339. |
| `maxAtime`    | int  | Maximum atime that can be parsed as RFC3339. |
| `minMtime`    | int  | Minimum mtime that can be parsed as RFC3339. |
| `maxMtime`    | int  | Maximum mtime that can be parsed as RFC3339. |
| `minCtime`    | int  | Minimum ctime that can be parsed as RFC3339. |
| `maxCtime`    | int  | Maximum ctime that can be parsed as RFC3339. |

#### Example

```yaml
name: file_time_validate
params:
  minMtime: 
    sourceType: value
    source: 2023-12-08T16:31:33Z
  maxMtime: 
    sourceType: value
    source: 2023-12-08T17:31:33Z
```

### exiftool_metadata_cleanup

Clear file metadata if possible (require exiftool).

#### Parameters

| Name                    | Type | Description                                          |
|-------------------------|------|------------------------------------------------------|
| `overwriteOriginalFile` | bool | Whether exiftool should overwrite the original file. |

#### Example

```yaml
name: exiftool_metadata_cleanup
params:
  overwriteOriginalFile: 
    sourceType: value
    source: false
```

### image_convert

Convert image to another format (require libvips).

#### Parameters

| Name         | Type   | Description                                                                   |
|--------------|--------|-------------------------------------------------------------------------------|
| `toMimeType` | string | MIME type of the output file.                                                 |
| `quality`    | string | Quality of the output file. Possible values: `low`, `medium`, `high`, `best`. |

#### Example

```yaml
name: image_convert
params:
  toMimeType: 
    sourceType: value
    source: image/jpeg
  quality: 
    sourceType: value
    source: high
```

### s3_upload

Upload file to S3-compatible storage.

#### Parameters

| Name              | Type    | Description        |
|-------------------|---------|--------------------|
| `accessKeyId`     | string  | Access key ID.     |
| `secretAccessKey` | string  | Secret access key. |
| `sessionToken`    | ?string | Session token.     |
| `region`          | string  | Storage region.    |
| `bucket`          | string  | Storage bucket.    |
| `endpoint`        | string  | Storage endpoint.  |

#### Example

```yaml
name: s3_upload
params:
  accessKeyId: 
    sourceType: secret
    source: aws_access_key_id
  secretAccessKey: 
    sourceType: secret
    source: aws_secret_access_key
  region: 
    sourceType: env_var
    source: AWS_REGION
  bucket: 
    sourceType: value
    source: my-bucket
  endpoint: 
    sourceType: env_var
    source: AWS_ENDPOINT
```

### command_exec

Execute arbitrary command.

The command parameters can be templated with the following variables:
* `{{.AbsolutePath}}` - absolute path to the file. Example: `/home/user/Videos/video.mp4`
* `{{.Filename}}` - filename with extension. Example: `video.mp4`
* `{{.Basename}}` - filename without extension. Example: `video`
* `{{.Extension}}` - file extension. Example: `.mp4`
* `{{.OriginalAbsolutePath}}` - absolute path to the original file. Example: `/home/user/Videos/video.mp4`
* `{{.OriginalFilename}}` - original filename with extension. Example: `video.mp4`
* `{{.OriginalBasename}}` - original filename without extension. Example: `video`
* `{{.OriginalExtension}}` - original file extension. Example: `.mp4`

#### Parameters

| Name                     | Type      | Description                                              |
|--------------------------|-----------|----------------------------------------------------------|
| `commandName`            | string    | Command name. Allows template vars.                      |
| `commandArgs`            | ?string[] | Command args. Allows template vars.                      |
| `outputFileDestination`  | ?string   | Path to the command's output file. Allows template vars. |
| `allowParallelExecution` | bool      | Whether the command can be executed in parallel.         |

#### Example

```yaml
name: command_exec
params:
  commandName:
    sourceType: value
    source: ffmpeg
  commandArgs:
    sourceType: value
    source: [
      "-i", "{{.AbsolutePath}}",
      "-c:v", "copy",
      "-c:a", "copy",
      "/tmp/{{.Basename}}.mp4",
    ]
  outputFileDestination:
    sourceType: value
    source: /tmp/{{.Basename}}.mp4
  allowParallelExecution: 
    sourceType: value
    source: false
```