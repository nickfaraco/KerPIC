<script>
  import { createEventDispatcher, onMount } from 'svelte';
  import { api } from '$lib/utils/api.js';

  export let photos = [];
  export let currentIndex = 0;
  export let visible = false;

  const dispatch = createEventDispatcher();

  let imageElement;
  let loading = true;
  let error = false;

  $: currentPhoto = photos[currentIndex];
  $: hasPrevious = currentIndex > 0;
  $: hasNext = currentIndex < photos.length - 1;

  // Preload adjacent images
  $: {
    if (currentPhoto) {
      preloadAdjacentImages();
    }
  }

  onMount(() => {
    // Handle keyboard events
    const handleKeydown = (event) => {
      if (!visible) return;

      switch (event.key) {
        case 'Escape':
          close();
          break;
        case 'ArrowLeft':
          event.preventDefault();
          previous();
          break;
        case 'ArrowRight':
          event.preventDefault();
          next();
          break;
      }
    };

    document.addEventListener('keydown', handleKeydown);
    
    return () => {
      document.removeEventListener('keydown', handleKeydown);
    };
  });

  function close() {
    visible = false;
    dispatch('close');
  }

  function previous() {
    if (hasPrevious) {
      currentIndex = currentIndex - 1;
      loading = true;
      error = false;
    }
  }

  function next() {
    if (hasNext) {
      currentIndex = currentIndex + 1;
      loading = true;
      error = false;
    }
  }

  function handleImageLoad() {
    loading = false;
    error = false;
  }

  function handleImageError() {
    loading = false;
    error = true;
  }

  function preloadAdjacentImages() {
    // Preload previous image
    if (hasPrevious) {
      const prevImg = new Image();
      prevImg.src = api.getThumbnailUrl(photos[currentIndex - 1].path, 800);
    }
    
    // Preload next image
    if (hasNext) {
      const nextImg = new Image();
      nextImg.src = api.getThumbnailUrl(photos[currentIndex + 1].path, 800);
    }
  }

  // Reset states when photo changes
  $: if (currentPhoto) {
    loading = true;
    error = false;
  }
</script>

{#if visible && currentPhoto}
  <!-- Overlay -->
  <div 
    class="fixed inset-0 flex items-center justify-center z-50"
    style="background-color: var(--bg-primary); opacity: 0.98;"
    on:click={close}
    role="dialog"
    aria-modal="true"
  >
    <!-- Main content -->
    <div class="relative w-full h-full flex items-center justify-center">
      <!-- Close button -->
      <button
        class="absolute top-4 right-4 w-10 h-10 rounded-full flex items-center justify-center z-10 transition-colors"
        style="background-color: var(--bg-secondary); color: var(--text-primary);"
        on:click={close}
        title="Close (Esc)"
      >
        ✕
      </button>

      <!-- Image counter -->
      <div class="absolute top-4 left-4 px-3 py-1 rounded text-sm z-10" 
           style="background-color: var(--bg-secondary); color: var(--text-primary);">
        {currentIndex + 1} / {photos.length}
      </div>

      <!-- Previous button -->
      {#if hasPrevious}
        <button
          class="absolute left-4 top-1/2 transform -translate-y-1/2 w-12 h-12 rounded-full flex items-center justify-center text-xl z-10 transition-colors hover:opacity-80"
          style="background-color: var(--bg-secondary); color: var(--text-primary);"
          on:click|stopPropagation={previous}
          title="Previous (←)"
        >
          ←
        </button>
      {/if}

      <!-- Next button -->
      {#if hasNext}
        <button
          class="absolute right-4 top-1/2 transform -translate-y-1/2 w-12 h-12 rounded-full flex items-center justify-center text-xl z-10 transition-colors hover:opacity-80"
          style="background-color: var(--bg-secondary); color: var(--text-primary);"
          on:click|stopPropagation={next}
          title="Next (→)"
        >
          →
        </button>
      {/if}

      <!-- Loading spinner -->
      {#if loading}
        <div class="absolute inset-0 flex items-center justify-center">
          <div class="animate-spin rounded-full h-12 w-12 border-b-2" style="border-color: var(--accent);"></div>
        </div>
      {/if}

      <!-- Error state -->
      {#if error}
        <div class="absolute inset-0 flex items-center justify-center">
          <div class="text-center" style="color: var(--text-primary);">
            <div class="text-4xl mb-4">⚠</div>
            <div class="text-lg">Failed to load image</div>
            <div class="text-sm mt-2" style="color: var(--text-secondary);">{currentPhoto.name}</div>
          </div>
        </div>
      {/if}

      <!-- Main image -->
      <div class="w-full h-full flex items-center justify-center p-4" on:click|stopPropagation>
        <img
          bind:this={imageElement}
          src={api.getThumbnailUrl(currentPhoto.path, 800)}
          alt={currentPhoto.name}
          class="max-w-full max-h-full object-contain"
          class:opacity-0={loading}
          class:opacity-100={!loading && !error}
          on:load={handleImageLoad}
          on:error={handleImageError}
        />
      </div>

      <!-- Photo info -->
      <div class="absolute bottom-4 left-4 right-4 p-4 rounded" 
           style="background-color: var(--bg-secondary); color: var(--text-primary);">
        <div class="flex items-start justify-between">
          <div class="flex-1 min-w-0">
            <div class="text-lg font-medium truncate">{currentPhoto.name}</div>
            <div class="text-sm mt-1 space-x-4" style="color: var(--text-secondary);">
              {#if currentPhoto.width && currentPhoto.height}
                <span>{currentPhoto.width} × {currentPhoto.height}</span>
              {/if}
              {#if currentPhoto.size}
                <span>{(currentPhoto.size / 1024 / 1024).toFixed(1)} MB</span>
              {/if}
              {#if currentPhoto.modTime}
                <span>{new Date(currentPhoto.modTime).toLocaleDateString()}</span>
              {/if}
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
{/if}

<style>
  /* Smooth transitions */
  img {
    transition: opacity 0.2s ease-in-out;
  }
</style>