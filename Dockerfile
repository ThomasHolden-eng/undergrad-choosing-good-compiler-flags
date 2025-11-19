FROM alpine:latest
COPY --from=golang:1.25.4-alpine /usr/local/go/ /usr/local/go/
 
ENV PATH="/usr/local/go/bin:${PATH}"
RUN apk update && apk add clang llvm
CMD ["/bin/sh"]