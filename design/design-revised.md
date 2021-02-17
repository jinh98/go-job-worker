Teleport coding challenge design document by Jin Huang

I) Architecture:

The library will consist of the worker and worker service. At a high level view, a worker will be able
to perform start, stop, and query status for a command, and the worker service provide a map 
for the client to have control over multiple workers. The worker service also provides the workers a
custom logger, which will be used to log output of a process (sdtout and stderr) during execution of a 
running job to a file or a io.Writer.

The structure will be as follows:

Worker:
- struct: UID, command, status, error, *logger, mutex
- functions: NewWorker(*logger, command), Start(), Stop(), Status(), execute(),
statusUpdate(new status), ExitStatus()

Worker Service:
- struct: workersmap, *logger, mutex
- functions: NewWorkerService(), GetWorker(UID), AddWorker(worker)

II) Functions design:

- Error handling will be done approprietly in each function.
- Locks with sync.mutex will be added where applicable

Worker:
- NewWorker(): take in a command and a *logger and creates a new worker
- Start(): Set status of worker as "running", and call Execute(). When Execute() returns finish or 
error, set status of worker accordingly.
- Stop(): Set status of a running worker as "killed" and terminates the process
- Status(): returns the status of a worker (with a lock)
- execute(): directly execute a command for a process, then pipe both Stdout and Stderr of the process
to stream process output to logger. When the process finishes or errors, return back to Start()
- statusUpdate(new status): updates the status of worker (with a lock)
- ExitStatus(): return the exit code of the worker's command if it ran

Worker Service:
- NewWorkerService(): creates a empty map of workers and a new logger whose pointer will be used 
for all workers to log process output to a file
- AddWorker(worker): inserts a worker into the map (with a lock)
- GetWorker(UID): returns worker (with a lock)

III) Tradeoffs and scope:
If not for simplicity, the following important features should be implemented:

1. In a scalable version of this project, the workers would ideally be backed with database and caching
    such that recovery is possible when a sudden failure is encountered.
2. Authentication and security are issues that needs to be addressed, which ideally an SSL handshake
    should be required to ensure secure communication b/t client and the server with https.
3. Clients that access this API would ideally be registered in a database with their credentials, with a
    client interface that accept authorized requests to protect server against malicious requests
4. Logs should be kept consistent and saved to enable referencing and debugging, could use a
    library like Logrus and send log files to a centralized logging platform.
5. Could use services like Docker to containerize the development environment for dependencies,
    and packages

IV) Edge cases consideration:

1. Too many concurrent running jobs with long execution time and high resource usage may cause
    problems (should not worry about in this current implementation)
2. Output of processes may not always be short or formatted to fit back to the client for output
    when requested (assume to have enough memory for output)
