# Stage 1. Build Binary File
FROM golang:1.16-alpine3.13 AS build
WORKDIR /app
COPY . . 
RUN go build -o main main.go 

# Stage 2. Copy and Execute Binary File without the rest of the files
FROM  alpine:3.13
WORKDIR /app
# copy binary from build stage
COPY --from=build /app/main .

EXPOSE 8080

CMD [ "/app/main" ]