#!/bin/sh

RETVAL=0

PACFILE="1.js"

DATA=""

for i in $(seq 1 100); do
    DATA="http://google.com/${i} google.com ${DATA}"
done

echo $DATA

./test_pac $PACFILE $DATA

exit $RETVAL
