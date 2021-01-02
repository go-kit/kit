FROM golang:1.15-buster as build

RUN curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | apt-key add -
RUN echo "deb http://dl.yarnpkg.com/debian/ stable main" | tee /etc/apt/sources.list.d/yarn.list
RUN apt-get update && apt-get install -y yarn sqlite3

RUN curl -sL https://deb.nodesource.com/setup_12.x | bash -
RUN apt-get install -y nodejs build-essential zip

WORKDIR /src
ENV HOME=/home
# Build Go server
ENV GO111MODULE on

RUN useradd --uid 1000 --gid 0 configr && \
    chown configr:root /src && \
    chmod g=u /src $HOME
USER 1000:0

COPY --chown=configr:root  . .

RUN go mod download

RUN cd ./ui/ && yarn install && yarn run build
RUN make build

######################################
# Copy from build to alpine image
######################################
FROM debian:buster-slim as production
RUN apt-get update && apt-get install -y ca-certificates

RUN useradd --uid 1000 --gid 0 configr

USER 1000:0

COPY --from=build  --chown=configr:root  /src/kit ./kit
COPY --from=build  --chown=configr:root  /src/ui/dist ./ui/dist


EXPOSE 9090


CMD ["sh", "-c", "./kit"]

