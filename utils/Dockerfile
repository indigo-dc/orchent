FROM scratch
ADD orchent /
ADD ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
ENTRYPOINT ["/orchent"]
CMD ["--help"]