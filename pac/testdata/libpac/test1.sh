#!/bin/sh

RETVAL=0

PACFILE="1.js"

. ./test_helper

test_proxy $PACFILE http://abcdomain.com abcdomain.com "Found proxy DIRECT"
test_proxy $PACFILE ftp://mydomain.com/x/ mydomain.com "Found proxy DIRECT"
test_proxy $PACFILE http://a.local/x/ a.local "Found proxy DIRECT"
test_proxy $PACFILE http://10.1.2.3/ 10.1.2.3 "Found proxy DIRECT"
test_proxy $PACFILE http://172.16.1.2/x/ 172.16.1.2 "Found proxy DIRECT"
test_proxy $PACFILE http://192.168.1.2/x/ 192.168.1.2 "Found proxy DIRECT"
test_proxy $PACFILE http://127.0.0.5/x/ 127.0.0.5 "Found proxy DIRECT"
test_proxy $PACFILE http://google.com/x google.com "Found proxy PROXY 4.5.6.7:8080; PROXY 7.8.9.10:8080"

exit $RETVAL
