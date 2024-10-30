# Builder builds golang binary, passes it to the plain alpine
FROM ${base_image}

WORKDIR /app

# Copy everything at the root level to build the binary
COPY . .
RUN go build -o /bin/attacker ./main.go

# Build docker with only server binary
FROM alpine:latest

RUN apk --no-cache add ca-certificates

# Copy golang binary from the builder on plain alpine image where Go is not installed (smaller image)
COPY --from=builder /bin/attacker /bin/
COPY ./test_cases/ /bin/test_cases/

CMD ["/bin/attacker"]
