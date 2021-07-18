const fs = require("fs").promises;
const path = require("path");
const esbuild = require("esbuild");
const sveltePlugin = require("esbuild-svelte");
const sassPlugin = require("esbuild-plugin-sass");
const aliasPlugin = require("esbuild-plugin-alias");
const preprocess = require("svelte-preprocess");

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
  const targetServerFile = path.join(
    __dirname,
    "/../build",
    folder,
    filename + ".server.js"
  );
  const targetClientFile = path.join(
    __dirname,
    "/../build",
    folder,
    filename + ".client.js"
  );

  await esbuild.build({
    entryPoints: [sourceFile],
    bundle: true,
    format: "esm",
    outfile: targetServerFile,
    plugins: [
      aliasPlugin({
        $lib: path.resolve(__dirname, "../example/src/lib"),
      }),
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
    logLevel: "info",
  });

  await iifeModule(targetServerFile);

  await esbuild.build({
    entryPoints: [sourceFile],
    bundle: true,
    format: "esm",
    target: "es2015",
    outfile: targetClientFile,
    minify: true,
    plugins: [
      aliasPlugin({
        $lib: path.resolve(__dirname, "../example/src/lib"),
      }),
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
    logLevel: "info",
  });
  await iifeModule(targetClientFile);
};

async function iifeModule(filename) {
  const esmModule = await fs.readFile(filename, "utf-8");
  let js = `(function () {\n${esmModule}})();`;
  js = js.replace(`export {`, "return {");
  js = js.replace(`;export{`, ";return{");
  js = js.replace(/([^\n,{}]+) as default/, "default: $1");

  await fs.writeFile(filename, js);
}
