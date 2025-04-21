---
authors: Richy HBM
---

# RFD - Remote Job Execution System

## What

The following document is a technical design to provide pre-implementation
details on the implementation of the level 4 software engineer remote job execution challenge

## Scope

The requirements for the Level 4 challenge, as taken from
[here](https://github.com/gravitational/careers/blob/main/challenges/systems/challenge-1.md#level-4)
are as follows:

### Library

* Worker library with methods to start/stop/query status and get the output of a job.
* Library should be able to stream the output of a running job.
  * Output should be from start of process execution.
  * Multiple concurrent clients should be supported.
* Add resource control for CPU, Memory and Disk IO per job using cgroups.

### API

* [GRPC](https://grpc.io) API to start/stop/get status/stream output of a running process.
* Use mTLS authentication and verify client certificate. Set up strong set of
  cipher suites for TLS and good crypto setup for certificates. Do not use any
  other authentication protocols on top of mTLS.
* Use a simple authorization scheme.

### Client

* CLI should be able to connect to worker service and start, stop, get status, and stream output of a job.

## Details

### CLI UX

The application will be a single binary application using subcommands to run the different functionality
Depending on whether this is ran as the server/api or the client, the behavior will change as follows:

#### Client

The client-side program subcommands will allow the main functionality, these will all require an argument
to identify the server endpoint:

- Start
   ```
   program start --server=<server endpoint> <command --arguments>
   ```

   This will then return a job ID that will be used in the remaining commands to reference this job,
   alternatively it may print an error message if the job was unable to start for any reason.

   The returning job ID will use UUIDs to ensure colliding IDs are not generated, using a library such as
   https://pkg.go.dev/github.com/google/UUID

- Stop
   ```
   program stop --server=<server endpoint> --job=<job ID>
   ```
   This will end the specified job, returning it's exit code. This will be a blocking call
   returning once the job has ended and returning to the user if it required calling Kill, and the jobs
   exit code.

   The stop process will take the `os.Process` calling `.Kill()` on it if it has not already ended.
   The return call will return the value in the variable `os.ProcessState.ExitCode`

   If the program had already ended at the point at which stop is called, the field
   `JobStopResponse.forceEnded` will be false, otherwise Kill will be called on the process and it will be true.

- Query
   ```
   program status --server=<server endpoint> --job=<job ID>
   ```
   This will be used to fetch details on a specific job, such as if the job is still running,
   if it exited cleanly, etc.

- Tail
   ```
   program tail --server=<server endpoint> --job=<job ID>
   ```
   The tail command will act in a similar way to the `tail -f` command, executing for the length
   of the job, and returning a full history of the jobs output. Passing in a previously ended job
   should return its full output history.

Ideally there would also be additional commands for things like fetching a list of jobs started by
the current user, as well as a history of jobs and their status, etc. But for the challenge this isn't
required.

#### API/Server

The server is intended to only work on linux/unix systems, as such if ran on a non compliant OS it should
return a message indicating as much. This can be done using the [runtime.GOOS](https://pkg.go.dev/runtime#GOOS) variable,
and check it is a linux based system. Ideally a flag could be added to force skip this check if you wanted to run on a different
compliant OS like freebsd, etc, but for simplicity, that will be skipped for this exercise.

CLI command should look like:

```
program serve --port=4567
```

### Authentication

Authentication for the program will be done using mTLS (TLS v1.3), with self signed x509 CA certificates.
The created certificates will be 256 bit elliptic-curve certificates using the sha256 hashing algorithm.
For simplicity these will all be bundled in to the executable directly

More info can be found in [known tradeoffs](#known-tradeoffs), on why this isn't optimal for production

### Authorization

Authorization for the program will use simple methods for identifying the user,
the common name field in the key's subject to get their username, this will then be sent up to the server
which will then look up the user and see what permissions they have access to.
These permissions will just be hard coded, but would ideally come from a config file or service

The permissions would look at the job command issued and check that the given user has permissions to
run a specific command, for this example that could be as simple as storing a list of strings on the
server and checking these against the passed up command.

As with the above section, more information can be found in [known tradeoffs](#known-tradeoffs) on improvements that could be made.

### API Specification

Please see the included `.proto` files in the [proto](../proto/) directory

### CGroups Restrictions

Resource restrictions for the server initiated jobs will be managed using cgroups v2, with the main limits
being placed on CPU usage, Memory usage, and disk IO usage. When a client sends a request to start a new job,
the server will need to generate a new cgroup directory for the job, keeping note of this to delete it at the end
and create the necessary file entries for the above limits.

For the sake of simplicity, these values will just be hardcoded and all jobs will have the same limits applied,
but each new job will be configured as its own cgroup.

Once the cgroup has been set up, the server will then execute the job using `SysProcAttr.CgroupFD`
to ensure the job is added to the cgroup from the start.

The proposed cgroup values will be as follows:

- cpu.max: 2 cores, 100% `200000 100000`
- memory.max: 1GB Max `1048576000`
- io.max: 1GB Read, 100MB Write `MAJOR:MINOR rbytes=1048576000 wbytes=10485760 rios=1000000 wios=1000000`

The IO limits should apply to all devices, as such entries should be added for each.

### Streaming To Multiple Clients

The streaming from host to host will be managed with gRPC, a new goroutine will be created for each new job,
starting it, and then capturing its output, when writing to memory a mutex will need to be used in order to
ensure there are no race conditions, etc.

When a client connects to retrieve the job output it will first need to read all existing data in memory, but
will then need to wait for any new data until the program has ended, this means that along with the output
a flag will also need to be kept so that any connected clients can know to end a connection.

To achieve this, a structure will need to exist that will keep check on the process still being ran and all
output being read, blocking the client if so, and unblocking only when new output is available or the job has ended. This will likely be done with a goroutine reading output and passing it to each of the clients
structure via channels, allowing the client to read the data as they are ready to receive it. For clients
connecting later, that means that before adding their unique channel to the list of channels to forward to,
the current up to date output will need to be pushed to it. The structure containing the channels will also need
to hold an indexing counter to keep track of what position each client has read up to, this would ensure if
a channel fills up before it can be fully read from, we are able to keep the correct position in the stream.

This structure will need to be per client so as to not have clients interfere with each other, it can also
manage reading the output data in a thread safe way and forwarding this to the clients via gRPC.

One consideration is that the process may run for a long period of time
causing excessive output and causing a large amount of memory to be consumed whilst storing this.
In an ideal scenario this data would be saved to disk once a job has finished executing (potentially even whilst
running) but for this exercise this will all be kept in memory.

### Possible Issues

- Command Injection

   Based on the above authorization design, it would be possible for a malicious user to run commands
   in an SQL-injection style attack, by running first a valid command and then appending the attack on to it

   For example: `ls /; rm -rf /*` `echo 1 && rm -rf /*` or using other shell based logic, to counter this,
   the program should stop taking input or commands after the pattern has been found, running just the
   initial segment, or alternatively error out informing the user they must change their command logic.

   To help mitigate this, arguments can be parametrized at the point at which the server executes the job,
   alternatively the server could also strip non alphanumeric characters, though for this challenge it will just
   use the parametrized approach

- No Command Checks

   The current system will allow any commands to be run, this means the user could issue an `rm` command
   or even download external programs and run malicious jobs, whilst this won't be accounted for in this
   challenge, this is something that would need to be thought about in order to guard against.

### Known Tradeoffs

- Certificates will use self-signed certificates and be built in to the executable, this would want to
use production certs managed in a more secure way, likely rotated periodically,
if actually deployed in production

- The server component should read authorization data from some form of config, mapping users -> job permissions,
for simplicity this will be hardcoded.
The job permission config would also include configuration for cgroup allowances.
This would likely be some form of configuration tool to manage jobs and user permissions centrally,
this could also group permissions to user tags, allowing users with a specific tag to have access to
all permissions linked, etc.

- For authorization the executable will fetch the current user, based on private/public
keys containing the user in one of their fields (with an additional override for testing/forcing specific users)

- Long running jobs could cause memory issues, ideally this would be stored on disk or in a 3rd party storage,
but for this challenge it will remain in memory