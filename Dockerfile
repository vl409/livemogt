FROM golang:1.22.1-bookworm AS builder

WORKDIR /build

ADD /microinit microinit/
RUN make -C microinit/

COPY /back/go.mod .
COPY /back/go.sum .
RUN go mod download

ADD /back/src  src/
COPY /back/makefile .

RUN make

FROM node:21-alpine3.18 as front_builder

RUN apk add make git

WORKDIR /front

ADD /front/static /front/static
ADD /front/js /front/js
COPY /front/makefile /front

RUN make init
RUN make tarball

FROM nginx:1.25.4-alpine-slim

COPY --from=builder /build/microinit/microinit /microinit
COPY --from=builder /build/bin/livemogt/ /livemogt
COPY --from=builder /build/bin/webmap /webmap

COPY /conf/nginx.conf /etc/nginx/nginx.conf
COPY --from=front_builder /front/build/app/dist/ /usr/share/nginx/html/

USER nobody

ENTRYPOINT ["/microinit", "/livemogt:/conf/livemogt_conf.json", "/webmap:/conf/webmap_conf.json", "/usr/sbin/nginx"]
