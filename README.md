# goscp - A Go SCP wrapper 

goscp supports recursive file and directory handling for both uploads and downloads.
 
goscp assumes that you've already established a successful SSH connection. 

## Examples

### Creating a client

```go
var sshClient *ssh.Client
    
// Connect with your SSH client
// ...

// Create a goscp client
c := goscp.NewClient(sshClient)

// Verbose output when communicating with host
// Outputs sent and received scp protocol messages to console
c.Verbose = true

// Show a progress bar for each file being sent or received
c.ShowProgressBar = true

// Each file being transferred has a progress bar output to console
// Customise the progress bar
c.ProgressBar.ShowSpeed = false
c.ProgressBar.Callback = func(output string) {
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
   log.Println(err)
   
   // Optionally, grab the entire stack of errors that occurred before failure
   log.Fatal(c.GetErrorStack())
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
c.Upload("~/Projects/goscp-src")
if c.GetLastError() != nil {
    log.Fatal(err)
}
```

### Cancellation

You can optionally (violently) cancel a download or upload in progress.

```go
c := goscp.NewClient(sshClient)
   
// Path on the remote machine
c.SetDestinationPath("/usr/local/src")

go func(){
    c.Upload("~/Projects/goscp-src")
    if err := c.GetLastError(); err != nil {
        log.Fatal(err)
    }
}()

// Cancel after 5 seconds
time.Sleep(time.Second * 5)
c.Cancel()
```

## License
BSD 3-Clause "New" License

## Related

* [How the SCP protocol works][oracle-scp-how] 
* [GB][gb]

[gb]: http://getgb.io/
[oracle-scp-how]: https://blogs.oracle.com/janp/entry/how_the_scp_protocol_works
