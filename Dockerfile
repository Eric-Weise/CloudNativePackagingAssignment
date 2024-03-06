# syntax=docker/dockerfile:1

FROM golang:1.21 AS build-stage

# Set destination for COPY
WORKDIR /app

# Copy files
COPY . .

#download dependencies
RUN go clean -modcache

RUN go mod download

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /voter-api


FROM alpine:latest AS run-stage

# JUST put in root
WORKDIR /

# Copy binary from build stage
COPY --from=build-stage /voter-api /voter-api

# Expose port
EXPOSE 1080

#set env variables.  Note for a container to get access to the host machine, 
#you reference the host machine by using host.docker.internal (at least in docker desktop)
ENV REDIS_URL=host.docker.internal:6379

# Run
CMD ["/voter-api"]
