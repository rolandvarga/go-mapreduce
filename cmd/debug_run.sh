#!/bin/zsh
ls -- {mr-[0-9]-[0-9],mr-out-*} | xargs -I_file rm _file
go run -race mrcoordinator.go *.txt
sort mr-out* | grep . > mr-wc-all
cat mr-wc-all | wc -l
