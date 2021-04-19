FROM node:alpine
# Create app directory
WORKDIR /usr/src/app
# Install app dependencies
# A wildcard is used to ensure both package.json AND package-lock.json are copied
COPY package*.json ./
RUN npm install --prod
COPY dist-prod .
# We're downgrading below root, so make sure any out folders are properly writable
USER node
CMD [ "node", "index.js" ]