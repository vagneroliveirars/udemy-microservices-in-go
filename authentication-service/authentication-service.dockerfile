# build a tiny docker image
FROM alpine:3.19.1

RUN mkdir /app

COPY authApp /app

CMD [ "/app/authApp" ]