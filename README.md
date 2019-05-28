# Goll

## Description
`Goll` is a project I created to try to learn golang.
`Goll` is a shell that is password protected, thus making it able to open a specific
 port and being accessed from any computer, no matter what the attacker's IP is.

It allows basic shell operations, open files, download files, and so on, but it is kinda basic, focusing on being small and mainly 
because I don't pay too much attention to its development, but any improvement is well accepted.

## MO
Its functionality is pretty simple:

A UDP port is listening, waiting for a password. If the password is correct, `Goll` opens a reverse shell towards
the other machine, allowing from then on code execution.
Due to how it works, I would not say it is neither a bind nor a reverse shell, but has a bit of both: 
Bind for authentication, reverse for code execution.

This is extremely useful in cases where your IP can change, or you have multiple computers from which you want to connect,
 but don't want to enable SSH (for whatever reason there might be).

## Installation

1. Clone this repo `git clone https://github.com/hamperfait/goll.git`
2. Change code or constants(if necessary)
3. Compile
    * `go build`
4. Run in victim's PC

## Constants

Following are the default ports that are being used:
* Remote UDP listening port: `6666`
* Local UDP bind port: `5555`
* Remote TCP file server port: `4444`
* Login password `cmd`

## Connection

1. Set up a listener using ncat or any other tool `ncat -lu 5555`
2. Connect to `Goll`: `ncat IP -u 6666`
    1. Enter the password `cmd`

`Goll` will connect to the listening netcat (the `5555` one).

## Compatibility

`Goll` has been mainly used on Windows, although Linux should work as well.

More work is planned, but no schedule will be provided.

## Bugs so far

The exec once worked, but not any longer apparently... I want to look into it, but there it goes.
