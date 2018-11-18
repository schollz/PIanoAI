
HASH=$(shell git describe --always)
LDFLAGS=-ldflags "-s -w -X main.Version=${HASH}"

build:
	go get -d -v github.com/jteeuwen/go-bindata/...
	go install -v github.com/jteeuwen/go-bindata/go-bindata
	rm -rf assets 
	cp -r static assets
	cd assets && gzip -9 -r *
	cp templates/index.html assets/index.html
	# minify static/css/rwtxt.css | gzip -9   > assets/rwtxt.css
	# minify static/css/normalize.css | gzip -9   > assets/normalize.css
	# minify static/css/dropzone.css | gzip -9  > assets/dropzone.css
	# minify static/js/rwtxt.js | gzip -9  > assets/rwtxt.js
	# # gzip -9 -c static/js/rwtxt.js > assets/rwtxt.js
	# minify static/js/dropzone.js | gzip -9 > assets/dropzone.js
	# minify static/css/prism.css | gzip -9 > assets/prism.css
	# minify static/js/prism.js | gzip -9  > assets/prism.js
	# gzip -9 -c static/img/logo.png  > assets/logo.png
	# cp -r static/img/favicon assets/
	# cd assets/favicon && gzip -9 *
	go-bindata -nocompress assets 
	go build -v ${LDFLAGS}
