const path = require("path");
const processDirectory = require("./processDirectory.cjs");

// @todo implement --watch
processDirectory(
  path.resolve(__dirname, "../example/src/routes"),
  path.resolve(__dirname, "../build/routes")
);
