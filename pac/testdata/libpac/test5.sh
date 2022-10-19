#!/bin/sh

RETVAL=0

PACFILE="4.js"

. ./test_helper

test_error $PACFILE http://foobar.example.com/x foobar.example.com "Javascript call failed: testing error handling"

exit $RETVAL
