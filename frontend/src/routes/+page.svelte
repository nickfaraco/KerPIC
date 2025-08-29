<script>
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';

  onMount(() => {
    // Redirect to gallery on load
    goto('/gallery');
  });
</script>

<svelte:head>
  <title>KerPIC - Photo Gallery</title>
</svelte:head>

<div class="min-h-screen bg-gray-900 text-white flex items-center justify-center">
  <div class="text-center">
    <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto"></div>
    <div class="mt-4 text-gray-400">Loading gallery...</div>
  </div>
</div>