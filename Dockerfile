FROM scratch

ADD ethr /

COPY .weave.yaml /
CMD ["/ethr", "-s", "-4", "-ip", "0.0.0.0", "-no"]
