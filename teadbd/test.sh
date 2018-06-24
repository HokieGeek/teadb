#!/bin/sh

curl -X PUT \
-v \
--header "Content-Type: application/json" \
http://localhost:8888/tea/42/entry \
--data '{"comments":"foobar","timestamp":"1/13/2012 22:11:36","date":"1/13/2012","time":"2200","rating":2,"pictures":null,"steeptime":"1m","steepingvessel_idx":8,"steeptemperature":180,"sessioninstance":"","sessionclosed":true,"fixins_list":null}'

curl -v http://localhost:8888/tea/42/entry
