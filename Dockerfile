FROM scratch

ADD ethr /

COPY .weave.yaml /
CMD ["/ethr -s"]
