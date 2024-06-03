FROM alpine:latest
LABEL maintainer="Rishi"
WORKDIR /src
COPY ./configs/.env ./configs/.env
COPY main /main
RUN chmod +x /main
EXPOSE 8000
ENTRYPOINT [ "/main" ]