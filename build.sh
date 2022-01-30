#!/usr/bin/env bash

platforms=("windows-amd64" "windows-386" "darwin-amd64" "darwin-arm64" "linux-386" "linux-amd64" "linux-arm64" "linux-arm")
tag=$(git for-each-ref --sort=-v:refname --format '%(refname:lstrip=2)' | grep -E "^v?[0-9]+\.[0-9]+\.[0-9]+$" | head -n1)

for i in ${platforms[@]}; do 

    goos=$(echo $i | awk -F '-' '{print $1}')
    goarch=$(echo $i | awk -F '-' '{print $2}')


    if [ $goos = "windows" ]; then
        env GOOS=$goos GOARCH=$goarch go build -o nextsubnet.exe cmd/nextsubnet/main.go
        zip -r nextsubnet-$tag-$goos-$goarch.zip nextsubnet.exe
        rm nextsubnet.exe
    else
        env GOOS=$goos GOARCH=$goarch go build -o nextsubnet cmd/nextsubnet/main.go
        tar -czvf nextsubnet-$tag-$goos-$goarch.tar.gz nextsubnet
        rm nextsubnet
    fi

done
