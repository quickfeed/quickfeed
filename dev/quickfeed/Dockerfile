FROM node:12.18.2
# FROM golang:1.15

WORKDIR /app

COPY ./package.json /app/package.json
COPY ./package-lock.json /app/package-lock.json
RUN npm ci

COPY . .

CMD ["node", "server.js"]
