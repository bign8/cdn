#! /bin/sh

rm data/*
mkdir -p data

for i in $(seq 1 2 60); do
  echo "----- Testing $i clients -----"
  TESTNUM=$i ./run.sh
done
