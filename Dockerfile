FROM acoshift/go-scratch

USER 65534:65534
COPY server /

EXPOSE 8080
EXPOSE 18080

ENTRYPOINT ["/server"]
