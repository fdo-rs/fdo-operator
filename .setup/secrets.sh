#!/bin/sh

oc new-project fdo-operator

for f in keys/*
do
    name=$(echo $f | awk -F'[/.]' '{ gsub("_", "-" ,$2); print $2 }')
    oc create secret generic "fdo-$name" --from-file=$f
done