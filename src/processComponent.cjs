const fs = require("fs").promises;
const path = require("path");
const esbuild = require("esbuild");
const sveltePlugin = require("esbuild-svelte");
const sassPlugin = require("esbuild-plugin-sass");

module.exports = async function processComponent(filename) {
  const folder = "routes";

  const sourceFile = path.resolve(
    __dirname,
    "..",
    "example",
    "src",
    folder,
    filename + ".svelte"
  );
  const targetFile = path.join(
    __dirname,
    "/../build",
    folder,
    filename + ".js"
  );

  await esbuild.build({
    entryPoints: [sourceFile],
    bundle: true,
    format: "esm",
    outfile: targetFile,
    plugins: [
      sassPlugin(),
      sveltePlugin({
        compileOptions: {
          generate: "ssr",
          sourcemap: false,
          css: false,
        },
      }),
    ],
    logLevel: "info",
  });

  const bundle = await fs.readFile(targetFile, "utf-8");

  const match = bundle.match(/export {\n  ([^}]+) as default\n};/);
  if (!match) {
    console.error("No default export detected");
  } else {
    const goModule =
      "(function () {" +
      bundle.replace(/export {[^}]+};/, "") +
      "\n  return {\n    default: " +
      match[1] +
      ",\n  };\n})();";
    await fs.writeFile(targetFile, goModule);
  }
};
