FROM scratch

ADD ethr /

COPY .weave.yaml /
CMD ["/ethr", "-s", "-no", "-4"]
