FROM alpine:3.23.2
COPY --from=golang:1.25.4-alpine /usr/local/go/ /usr/local/go/
 
ENV PATH="/usr/local/go/bin:${PATH}"
RUN apk update && apk add clang21 llvm
CMD ["/bin/sh", "-c", "go mod tidy && exec /bin/sh"]