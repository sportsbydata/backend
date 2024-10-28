FROM golang:1.23-alpine AS build

ARG BINARY
ARG ARCH
ARG OS

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=${ARCH} GOOS=${OS} go build -o ${BINARY} ./cmd/${BINARY}

FROM public.ecr.aws/lambda/provided:al2023

ARG BINARY

WORKDIR /opt

COPY --from=build /build/${BINARY} ./main

ENTRYPOINT ["./main"]
