ifconfig.co: Simple IP address lookup service
=============================================

A simple service for looking up your IP address. This is the code that powers
http://ifconfig.co

Basic usage
===========

The usual suspects
------------------
    $ curl ifconfig.co
    127.0.0.1
    $ wget -q -O - ifconfig.co
    127.0.0.1

BSD fetch
---------
    $ fetch -q -o - ifconfig.co
    127.0.0.1

Pass the appropriate flag (usually -4 and -6) to your tool to switch between 
IPv4 and IPv6 lookup.

Features
========
* Easy to remember domain name
* Supports IPv4 and IPv6
* Open source
* Fast
* Supports typical CLI tools (curl, wget and fetch)

Dependencies (optional)
=======================
* Daemonize script depends on https://github.com/bmc/daemonize

Why?
====
* To scratch an itch
* An excuse to use Go for something
* Faster than ifconfig.me and has IPv6 support

Code style
==========
    $ gofmt -tabs=false -tabwidth=4
