<script>
  import Counter from "../lib/Counter.svelte";
  import "../global.scss";

  const now = new Date();
  $: milliseconds = (now.getMilliseconds() / 1000).toFixed(3).substr(2);
</script>

<h1>Hallo</h1>
<p>
  The time is: {now.getHours()}:{now.getMinutes()}:{now.getSeconds()}.{milliseconds}
</p>

<Counter />

<style>
  h1 {
    font-size: 10px;
  }
</style>
