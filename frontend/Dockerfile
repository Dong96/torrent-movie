FROM node:latest as builder
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

FROM nginx
WORKDIR /app
COPY --from=builder /app/dist/ .
COPY nginx.conf /etc/nginx

EXPOSE 8083