<script setup lang="ts">
import { RouterView } from "vue-router";
import { Show, SignInButton, SignUpButton, UserButton } from "@clerk/vue";
</script>

<template>
  <RouterView />
  <header>
    <Show when="signed-out">
      <SignInButton />
      <SignUpButton />
    </Show>
    <Show when="signed-in">
      <UserButton />
    </Show>
  </header>
</template>

<style>
* {
  box-sizing: border-box;
}

body {
  margin: 0;
  font-family:
    -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue",
    Arial, sans-serif;
}
</style>
