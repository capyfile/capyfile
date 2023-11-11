This example demonstrates how you can build a simple file uploading endpoint using `capysvr`. It
shows how to proxy the requests to `capysvr` and use its configuration capabilities to control the
uploading process.

To try it out:
```bash
# Run capysvr that will handle the file uploading (need Docker)
$ ./capysvr.sh

# Run the proxy that will forward the requests to capysvr
$ go run animal_saver.go

# Upload a file
$ curl -X POST -F "file=@$HOME/Pictures/capybara.png" http://localhost:8080/capybaras
```
