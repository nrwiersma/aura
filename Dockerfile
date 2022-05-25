FROM gcr.io/distroless/static:nonroot

COPY aura /aura

EXPOSE 8080
CMD ["/aura", "server"]
