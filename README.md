
# go-job-worker

  

This project provides a worker library that can start, query, and stop an arbitary Linux process and get the output of the process.

  

## Requirements

This project runs on a Linux environment and uses go 1.15. There are no external libraries used in this implementation.

  

## Building the API

 `git clone` into this respository and do `cd go-job-worker`

 To build the project, run:

```bash

$ go build

```

  

This command will build the project and create an executable`go-job-worker` for the sample program in `main.go` to demonstrate basic uses of the library. 

  
  

## Using the sample program

```bash

$ ./go-job-worker

```

This command will run the executable for the sample program.

  
  

## Running tests

  
Some test cases are provided to demonstrate correctness and functionality of this worker library (but not full coverage).

  

```bash

go test ./jobworker

```

