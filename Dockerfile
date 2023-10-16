FROM golang:1.21.1 AS builder
LABEL stage=screenshotbuilder

WORKDIR /build/screenshot

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /app/screenshot cmd/server.go

FROM chromedp/headless-shell:latest
RUN apt-get update
RUN apt-get install -y dumb-init wget
RUN wget http://ftp.uk.debian.org/debian/pool/contrib/m/msttcorefonts/ttf-mscorefonts-installer_3.8.1_all.deb
RUN apt-get install -y ttf-wqy-microhei ttf-wqy-zenhei xfonts-wqy cabextract
RUN dpkg -i ttf-mscorefonts-installer_3.8.1_all.deb
RUN apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* && fc-cache -f

WORKDIR /app

COPY --from=builder /app/screenshot /app/screenshot

ENTRYPOINT ["dumb-init" , "--", "./screenshot"]
