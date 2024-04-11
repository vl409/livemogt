OPTS=
OPTS=--progress=plain --no-cache

IMAGENAME=livemogt:0.8
HOSTCONF=./conf
HOSTDIR=./var
PORT=8080

.PHONY: livemogt livemogt-run

livemogt:
	cp conf/track.gpx front/static
	docker build -f Dockerfile $(OPTS) --tag $(IMAGENAME) .

livemogt-run:
	docker run --rm -v $(HOSTCONF):/conf -v $(HOSTDIR):/var -p $(PORT):$(PORT) $(IMAGENAME)

clean:
	@rm -f front/static/track.gpx

