all: tarball

init:
	mkdir -p build && cd build && npm create ol-app app

build/app/index.html: static/index.html
	cp $^ $@

build/app/style.css: static/style.css
	cp $^ $@

build/app/map.js: js/map.js
	cp $^ $@

build/app/vite.config.js: js/vite.config.js
	cp $^ $@

.PHONY assets: build/app/index.html \
               build/app/style.css \
               build/app/map.js \
               build/app/vite.config.js
	cd build/app && npm run build

tarball: assets
	cp static/track.gpx build/app/dist/assets
	cd build/app/dist && tar cvzf livemogt-web.tar.gz * && cp livemogt-web.tar.gz ../../

clean:
	rm -rf app/dist/
	rm -f livemogt-web.tar.gz
