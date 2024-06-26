# Rocked

## A simple utility to handle container

A tool built while I learn about containers and how they work.

This is based on the [rubber-docker](https://github.com/Fewbytes/rubber-docker) project.

## Usage

You can run the command by either compiling it first:
```
# go build .
```
And then run it as:
```
# ./rocked
A simple utility to handle container.

Usage:
  rocked [command]

Available Commands:
  help        Help about any command
  run         Runs a process

Flags:
  -h, --help      help for rocked
  -v, --verbose   Enable verbose logging

Use "rocked [command] --help" for more information about a command.
```
The `run` command takes:
```
# ./rocked run -h
Runs a process

Usage:
  rocked run [flags]

Flags:
  -e, --env stringArray   Sets environment variables. It can be repeated
  -h, --help              help for run
  -i, --image string      Use the container image (default "Fedora")

Global Flags:
  -v, --verbose   Enable verbose logging
```

To pass a command to run, use the double `-` to mark the end of the command options:
```
# sudo ./rocked run -i Fedora -- /usr/bin/whoami
```

The `-i` flag is mandatory. For now, the path where the images should be placed is `/tmp/test-chroot`.
The program will be then use the path plus the image name, for example, `/tmp/test-chroot/Fedora`.

For now, you can use the script in `utils` named `prep_chroot.sh` to download a Fedora or an Ubuntu container image in the current directory.
Once downloaded, extract the archive to `/tmp/test-chroot/{fedora, ubuntu}`.

## Levels

[Here](doc/LEVELS.md) are some notes on the various levels.
