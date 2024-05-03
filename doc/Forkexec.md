# Fork and Exec

Fork and exec is a very common way of "splitting" the process in two parts where one executes a separate program.
This is done by using the system calls [fork()](https://www.man7.org/linux/man-pages/man2/fork.2.html) and [exec()](https://man7.org/linux/man-pages/man3/exec.3.html).

After executing fork, the original process is duplicated to create an exact copy and both keep running from the same point after the fork
system call.
The code needs to check the return code of the fork system call and, based on that (return of 0 means it's the child, return of a pid number is the parent process),
it will either wait (with the system call wait()) for the child process to change of status or, in the case of the child, load another executable which will be in the child process.

To implement this in Golang, I took inspiration on how this was done in the package [syscall](https://pkg.go.dev/syscall).
However, the approach taken there is to combine fork and exec into one single function call (https://cs.opensource.google/go/go/+/master:src/syscall/exec_unix.go;l=143?q=forkExec&sq=&ss=go%2Fgo).
I decided to keep the traditional C approach of having a separate function for fork() and for exec().

The only issue I have (and I'm not yet sure if this is how Golang is handling the various threads) is that troubleshooting this with the usual *nix tools becomes more difficult, in particular with
`strace`.
