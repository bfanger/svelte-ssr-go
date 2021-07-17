# Svelte SSR Go

Using the http server from [go](https://golang.org/) and combining it with the ssr template engine from [svelte](https://svelte.dev/).

An experiment to see if it was possible (poc for a c# implementation)
And a challenge to see if v8go could beat node in a benchmark.

# Setup

```
yarn install
yarn build # compiles svelte into js for go
yarn dev # starts go in watch mode
```

# Implemented

- Rendering a SSR component from Go

# Todo

- sveltekit routing
- sveltekit load
- polyfills for browser api's
- polyfill common node modules
- hydrate
- build --watch (or vite)
