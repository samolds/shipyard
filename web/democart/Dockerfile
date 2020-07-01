FROM node:13.12.0-alpine AS builder

WORKDIR /app
COPY package.json package-lock.json ./

ENV PATH /app/node_modules/.bin:$PATH
RUN npm ci --silent && npm install react-scripts@3.4.1 -g --silent

COPY . .
RUN npm run build


FROM nginx:stable-alpine
COPY --from=builder /app/build /usr/share/nginx/html

CMD ["nginx", "-g", "daemon off;"]
