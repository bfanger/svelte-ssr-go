const fs = require("fs").promises;
const path = require("path");
const { execSync } = require("child_process");

const { compile } = require("svelte/compiler");

module.exports = async function processComponent(filename) {
  const folder = "pages";

  const sourceFile = path.resolve(
    __dirname,
    "../example",
    folder,
    filename + ".svelte"
  );
  const intermediateFile = sourceFile + ".js";
  const targetFile = path.join(
    __dirname,
    "/../build",
    folder,
    filename + ".js"
  );

  const source = await fs.readFile(sourceFile, "utf-8");
  const result = compile(source, { filename, generate: "ssr", css: false });

  for (const warning of result.warnings) {
    console.warn(warning);
  }
  await fs.writeFile(intermediateFile, result.js.code);
  const esbuildCmd =
    __dirname +
    "/../node_modules/.bin/esbuild " +
    "--bundle " +
    "--format=esm " +
    intermediateFile +
    " --outfile=" +
    targetFile;

  execSync(esbuildCmd, { encoding: "utf8" });

  const bundle = await fs.readFile(targetFile, "utf-8");
  const match = bundle.match(/export {\n  ([^}]+) as default\n};/);
  if (!match) {
    console.error("No default export detected");
  } else {
    const goModule =
      bundle.replace(/export {[^}]+};/, "") +
      "\n" +
      match[1] +
      "; // export to golang";
    await fs.writeFile(targetFile, goModule);
    await fs.unlink(intermediateFile);
  }
};
