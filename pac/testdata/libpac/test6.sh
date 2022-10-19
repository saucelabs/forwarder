#!/bin/sh

RETVAL=0

PACFILE="5.js"

. ./test_helper

test_error $PACFILE http://foobar.example.com/x foobar.example.com "Javascript call failed: Error: testing error handling"

exit $RETVAL
