# goscp - A Go SCP Client 

goscp supports recursive file and directory handling for both uploads and downloads.
 
goscp assumes that you've already established a successful SSH connection. 

## Examples

### Creating a client

```go
var sshClient *ssh.Client
    
// Connect with your SSH client
// ...

c := goscp.NewClient(sshClient)

// Verbose output when communicating with host
// Outputs sent and received scp protocol messages to console
c.Verbose = true

// Each file being transferred has a progress bar output to console
c.ProgressCallback = func(output string) {
    // This allows you to control what will happen instead 
}

```

### Downloading
   
```go
c := goscp.NewClient(sshClient)

// Path on your local machine 
c.SetDestinationPath("~/Downloads")

// Path on the remote machine
// Supports both files and directories
c.Download("/var/www/media/images")
if c.GetLastError() != nil {
   log.Fatal(err)
}
```

### Uploading

```go
c := goscp.NewClient(sshClient)

// Path on the remote machine
c.SetDestinationPath("/usr/local/src")

// Stop on local FS errors that occur during filepath.Walk
c.StopOnOSError = true

// Path on your local machine
// Supports both files and directories
c.Upload("~/Documents/media/videos")
if c.GetLastError() != nil {
    log.Fatal(err)
}
```

## License
BSD 3-Clause "New" License

## Related

* [How the SCP protocol works][oracle-scp-how] 
* [GB][gb]

[gb]: http://getgb.io/
[oracle-scp-how]: https://blogs.oracle.com/janp/entry/how_the_scp_protocol_works
