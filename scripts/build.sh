VERSION="0.0.3"
rm -r "../builds/"
mkdir "../builds"
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ../builds/1PanelHelper-$VERSION-Darwin-amd64 ../
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ../builds/1PanelHelper-$VERSION-Windows-amd64.exe ../
CGO_ENABLED=0 go build -o ../builds/1PanelHelper-$VERSION-Linux-amd64 ../