https://medium.com/go-walkthrough/go-walkthrough-io-package-8ac5e95a9fbd#.d2ebstv0q


1. Implement a curl-like CLI tool (let’s name it curly) that downloads a file from the provided URL. By default, curly sends the output to /dev/null and sets exit code to 0 if download was successful, 1 if failed + prints error to stderr.
   But if -output=FILE flag is set - tool should store the contents of URL to the FILE. If “-output=-” tool should print output to stdout
2. Add flag -md5, so if set md5 sum will be printed to Stderr
3. Add flag -output-chunked FILEPREFIX - if set - content should be splitted to 3.5 Mb files FILEPREFIX.0 FILEPREFIX.1 ... FILEPREFIX.N so it could be stored on floppy disks :slightly_smiling_face:. output-chunked should work together with other flags!



```shell
wget https://play.golang.org/p/HmnNoBf0p1z; md5sum HmnNoBf0p1z
# 474b18855ceed917e30c29b98dcc1854

go run main.go --md5=true https://play.golang.org/p/HmnNoBf0p1z
```


```shell
go build
./curly --output=foobig --output-chunked=foo --md5=true https://i.redd.it/dujlhm3dqh951.png

# d41d8cd98f00b204e9800998ecf8427e

cat foo.0 foo.1 foo.2 foo.3 foo.4 foo.5 foo.6 foo.7 foo.8 foo.9 foo.10 foo.11 foo.12 > mergedfoo
# md5sum mergedfoo
b4be6a7103e47d6f8fe247d66d797bbd 

#md5sum foobig
b4be6a7103e47d6f8fe247d66d797bbd  foobig


```