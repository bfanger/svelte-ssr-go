import path from "path";
import { Transform } from "stream";
import { src, dest, watch, parallel } from "gulp";
import { createGulpEsbuild } from "gulp-esbuild";
import livereload from "gulp-livereload";
import sveltePlugin from "esbuild-svelte";
import sassPlugin from "esbuild-plugin-sass";
import preprocess from "svelte-preprocess";
import type { BufferFile } from "vinyl";

const gulpEsbuild = createGulpEsbuild({});
const base = path.resolve(__dirname, "example/src/routes") + "/";

function buildServer() {
  return src(base + "**/*.svelte")
    .pipe(
      gulpEsbuild({
        logLevel: "warning",
        format: "esm",
        bundle: true,
        plugins: [
          sassPlugin(),
          sveltePlugin({
            preprocess: preprocess(),
            compileOptions: {
              generate: "ssr",
            },
          }),
        ],
      })
    )
    .pipe(iifeModule())
    .pipe(dest("./build/server/"));
}
function buildClient() {
  return src(base + "**/*.svelte")
    .pipe(
      gulpEsbuild({
        logLevel: "warning",
        format: "esm",
        bundle: true,
        minify: true,
        plugins: [
          sassPlugin(),
          sveltePlugin({
            preprocess: preprocess(),
            compileOptions: {
              generate: "dom",
              hydratable: true,
            },
          }),
        ],
      })
    )
    .pipe(iifeModule())
    .pipe(dest("./build/client/"))
    .pipe(livereload());
}

const buildTask = parallel([buildServer, buildClient]);

function watchTask() {
  livereload.listen();
  watch(["example/**/*.{svelte,ts,js}"], buildTask);
}

exports.build = buildTask;
exports["build:server"] = buildServer;
exports["build:client"] = buildClient;
exports.watch = watchTask;

function iifeModule() {
  const transformStream = new Transform({ objectMode: true });

  transformStream._transform = function (file: BufferFile, encoding, callback) {
    if (file.basename.match(/.js$/)) {
      const esmModule = file.contents.toString(encoding);
      const regex = /export([ ]?{)([^}]+)}/;
      const match = esmModule.match(regex);
      if (!match) {
        console.warn("no export block detected in " + file.basename);
      } else {
        const converted =
          "return" +
          match[1] +
          match[2].replace(/([^\n,{} ]+) as ([^;,\n }]+)/g, "$2: $1") +
          "}";
        file.contents = Buffer.from(
          `(function () {\n${esmModule.replace(regex, converted)}})();`,
          encoding
        );
      }
    }
    callback(null, file);
  };

  return transformStream;
}
