const fs = require("fs").promises;
const path = require("path");
const esbuild = require("esbuild");
const sveltePlugin = require("esbuild-svelte");
const sassPlugin = require("esbuild-plugin-sass");
const preprocess = require("svelte-preprocess");

module.exports = async function processComponent(sourceFilename, targetDir) {
  const filename = path.basename(sourceFilename).replace(/\.svelte$/, "");
  const targetServerFile = path.join(targetDir, filename + ".server.js");
  const targetClientFile = path.join(targetDir, filename + ".client.js");

  await esbuild.build({
    entryPoints: [sourceFilename],
    bundle: true,
    format: "esm",
    outfile: targetServerFile,
    plugins: [
      sassPlugin(),
      sveltePlugin({
        preprocess: preprocess(),
        compileOptions: {
          generate: "ssr",
          sourcemap: false,
          css: false,
        },
      }),
    ],
    logLevel: "warning",
  });

  await iifeModule(targetServerFile);

  await esbuild.build({
    entryPoints: [sourceFilename],
    bundle: true,
    format: "esm",
    target: "es2015",
    outfile: targetClientFile,
    minify: true,
    plugins: [
      sassPlugin(),
      sveltePlugin({
        preprocess: preprocess(),
        compileOptions: {
          generate: "dom",
          hydratable: true,
          css: false,
        },
      }),
    ],
    logLevel: "warning",
  });
  await iifeModule(targetClientFile);
};

async function iifeModule(filename) {
  const esmModule = await fs.readFile(filename, "utf-8");
  const regex = /export([ ]?{)([^}]+)}/;
  const match = esmModule.match(regex);
  if (!match) {
    console.warn("no export block detected in " + filename);
    return;
  }
  const converted =
    "return" +
    match[1] +
    match[2].replace(/([^\n,{} ]+) as ([^;,\n }]+)/g, "$2: $1") +
    "}";

  await fs.writeFile(
    filename,
    `(function () {\n${esmModule.replace(regex, converted)}})();`
  );
}
