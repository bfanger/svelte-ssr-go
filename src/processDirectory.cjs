const processComponent = require("./processComponent.cjs");
const fs = require("fs").promises;
const path = require("path");

module.exports = async function processDirectory(dir, out) {
  const processing = [];
  for (const entry of await fs.readdir(dir)) {
    const fullpath = path.join(dir, entry);
    if (entry.endsWith(".svelte")) {
      processing.push(processComponent(fullpath, out));
    } else if (await isDirectory(fullpath)) {
      processing.push(processDirectory(fullpath, path.join(out, entry)));
    }
  }
  await Promise.all(processing);
};

async function isDirectory(dir) {
  const stat = await fs.stat(dir);
  return !!stat.isDirectory();
}
