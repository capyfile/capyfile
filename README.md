<p align="center">
  <img height="50%" width="50%" src="capybara.png" alt="Capyfile logo">
</p>

**Capyfile** - highly customizable file processing pipeline with built-in HTTP server, CLI,
and worker interfaces.

What we are pursuing here:
* Easy setup
* High customization
* Wide range of file processing operations

## How to use?

File processing pipeline can be set up in **two simple steps**.

#### Step 1: Write configuration file that describes the file processing pipeline

```yaml
---
version: '1.2'
name: photos
processors:
  - name: archive
    operations:
      # read the files from the directory
      - name: filesystem_input_read
        params:
          target:
            sourceType: value
            source: "/home/user/Photos/*"
      # check the file type
      - name: file_type_validate
        params:
          allowedMimeTypes:
            sourceType: value
            source:
              - image/jpeg
              - image/x-canon-cr2
              - image/heic
              - image/heif
      # if the file type is right, upload the file to S3
      - name: s3_upload
        targetFiles: without_errors
        params:
          accessKeyId:
            sourceType: env_var
            source: AWS_ACCESS_KEY_ID
          secretAccessKey:
            sourceType: env_var
            source: AWS_SECRET_ACCESS_KEY
          endpoint:
            sourceType: value
            source: "s3.amazonaws.com"
          region:
            sourceType: value
            source: "us-east-1"
          bucket:
            sourceType: env_var
            source: AWS_PHOTOS_BUCKET
      # if the file type is right, and it is successfully uploaded to S3,
      # remove the file from the filesystem
      - name: filesystem_input_remove
        targetFiles: without_errors
```

Both YAML and JSON formats are supported.

#### Step 2: Run the file processing pipeline

```bash
# set the environment variables if you use any
export AWS_ACCESS_KEY_ID=...
export AWS_SECRET_ACCESS_KEY=...
export AWS_PHOTOS_BUCKET=...

# run the file processing pipeline via capycmd command line application,
# enable concurrency to make it faster
capycmd -f photos.pipeline.yml --concurrency photos:archive
```

That's it 🎉

## Want to know more?

* [Operations](docs/operations.md)
* [Runners](docs/runners.md)
* [Performance](docs/performance.md)
* [Examples](docs/examples.md)
* [Development](docs/development.md)

Also check out the following articles:
* [Do whatever you want with your files, and do it quickly](https://dev.to/sky003/do-whatever-you-want-with-your-files-and-do-it-quickly-4od2)
* [Integrate any command into your file-processing pipeline](https://dev.to/sky003/integrate-any-command-into-your-file-processing-pipeline-3jbh)

## Why?

This project was created to with the following goals in mind:
1. Reduce the amount of boilerplate code and custom scripts.
2. Cover as many use cases as possible.
3. Provide declarative way to describe file processing pipelines.

