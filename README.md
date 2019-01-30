# gpuoff

A nvidia gpu monitor that shuts down a VM `n` minutes after GPU idle.

## Why?

I was incuring too much $$ for my company by forgetting to turn off my GPU based VMs.

This keeps me out of trouble.

## Install

This uses a `syscall.Reboot` which requires elevated permission. Running under `sudo` with `go` does interesting things due to the super-user environment. To avoid this, build the project and run the binary as `sudo`.

```
go get github.com/sabhiram/gpuoff
cd $GOPATH/src/github.com/sabhiram/gpuoff
go build .
```

This should leave you with a binary named `gpuoff`.

## Usage

```
Usage of gpuoff:
  -[i]gnore     : list of string regular expressions
        list of processes to ignore (always on like XORG)
  -i[n]terval   : duration
        duration between checking GPU status (default 10s)
  -[t]imeout    : duration
        duration to shutdown after GPU(s) are idle (default 15m0s)
```
