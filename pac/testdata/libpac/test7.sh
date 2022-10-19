#!/bin/sh

RETVAL=0

PACFILE="6.js"

. ./test_helper

test_proxy $PACFILE http://this.domain.does.not.exist/ this.domain.does.not.exist "Found proxy PROXY :8080"

exit $RETVAL
