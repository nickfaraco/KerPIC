<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import ComparisonView from '$lib/components/ComparisonView.svelte';
  import { selectedImages, selectedPhotos } from '$lib/stores/app.js';

  let mounted = false;

  onMount(() => {
    mounted = true;
    
    console.log('Compare page mounted, selected images:', $selectedImages.length);
    
    // If not enough images, redirect back to gallery
    if ($selectedImages.length < 2) {
      console.log('Not enough images, redirecting back to gallery');
      goto('/gallery');
    }
  });

  function handleComparisonExit() {
    // Clear both selectedImages and selectedPhotos stores and go back to gallery
    selectedImages.set([]);
    selectedPhotos.set(new Set());
    console.log('Cleared selection state after comparison exit');
    goto('/gallery');
  }
</script>

<svelte:head>
  <title>Compare Photos - KerPIC</title>
</svelte:head>

<div class="h-full">
  {#if mounted && $selectedImages.length >= 2}
    <ComparisonView on:exit={handleComparisonExit} />
  {/if}
</div>