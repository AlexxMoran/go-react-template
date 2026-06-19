FROM node:26-alpine

WORKDIR /frontend

COPY package*.json ./
RUN npm install

COPY . .

EXPOSE 4200

CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]
