#!/bin/bash

SELF="$(pwd)/$0"
DIR=`dirname $SELF`

pushd $DIR
XDG_CONFIG_HOME=. terminator --layout=FourByFour
popd
