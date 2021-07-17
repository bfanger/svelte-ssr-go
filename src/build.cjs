const processComponent = require("./processComponent.cjs");
const fs = require("fs").promises;
const path = require("path");

async function main() {
  const entries = await fs.readdir(
    path.resolve(__dirname, "../example/src/routes")
  );
  const processing = [];
  for (const entry of entries) {
    if (entry.endsWith(".svelte")) {
      processing.push(processComponent(entry.replace(/\.svelte$/, "")));
    }
  }
  await Promise.all(processing);
}
main();
// @todo implement --watch
