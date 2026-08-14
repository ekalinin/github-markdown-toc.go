FROM gcr.io/distroless/static:nonroot

COPY gh-md-toc /gh-md-toc

ENTRYPOINT ["/gh-md-toc"]
