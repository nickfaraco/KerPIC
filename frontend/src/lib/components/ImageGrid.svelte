<script>
  import { onMount } from 'svelte';
  import { api } from '$lib/utils/api.js';
  import { currentView, selectedFolder, folderImages, selectedImages } from '$lib/stores/app.js';
  
  let loading = true;
  let error = null;
  let selectedIndices = new Set();

  $: images = $folderImages;

  onMount(async () => {
    if (!$selectedFolder) {
      currentView.set('browse');
      return;
    }

    try {
      const imageList = await api.getImages($selectedFolder.path);
      folderImages.set(imageList);
      loading = false;
    } catch (err) {
      error = err.message;
      loading = false;
    }
  });

  function toggleSelect(image, index) {
    if (selectedIndices.has(index)) {
      selectedIndices.delete(index);
    } else {
      selectedIndices.add(index);
    }
    selectedIndices = selectedIndices; // Trigger reactivity

    // Update the store
    const selected = images.filter((_, i) => selectedIndices.has(i));
    selectedImages.set(selected);
  }

  function startComparison() {
    if ($selectedImages.length < 2) {
      alert('Please select at least 2 images to compare.');
      return;
    }
    currentView.set('compare');
  }

  function selectAll() {
    if (selectedIndices.size === images.length) {
      selectedIndices.clear();
      selectedImages.set([]);
    } else {
      selectedIndices = new Set(images.map((_, i) => i));
      selectedImages.set([...images]);
    }
  }

  function handleKeydown(event) {
    switch (event.key) {
      case 'Escape':
        event.preventDefault();
        currentView.set('browse');
        break;
      case 'Enter':
        event.preventDefault();
        startComparison();
        break;
      case 'a':
      case 'A':
        event.preventDefault();
        selectAll();
        break;
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="h-full flex flex-col">
  <!-- Header -->
  <div class="p-6 border-b border-gray-700 flex items-center justify-between">
    <div>
      <h2 class="text-xl font-semibold text-white mb-1">{$selectedFolder?.name}</h2>
      <p class="text-gray-400 text-sm">Select images to compare. Click images to select them, then press Enter to start comparing.</p>
    </div>
    
    <div class="flex items-center gap-4">
      {#if $selectedImages.length > 0}
        <span class="text-sm text-blue-400">{$selectedImages.length} selected</span>
        <button
          class="btn-primary"
          disabled={$selectedImages.length < 2}
          on:click={startComparison}
        >
          Compare ({$selectedImages.length})
        </button>
      {/if}
      <button class="btn-secondary" on:click={() => currentView.set('browse')}>
        ← Back
      </button>
    </div>
  </div>

  <!-- Content -->
  <div class="flex-1 overflow-auto">
    {#if loading}
      <div class="flex items-center justify-center h-64">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
        <span class="ml-3 text-gray-400">Loading images...</span>
      </div>
    {:else if error}
      <div class="flex items-center justify-center h-64">
        <div class="text-center">
          <div class="text-red-400 mb-2">⚠ Error loading images</div>
          <div class="text-gray-400 text-sm">{error}</div>
        </div>
      </div>
    {:else if images.length === 0}
      <div class="flex items-center justify-center h-64">
        <div class="text-center">
          <div class="text-gray-400 mb-2">📷 No images found</div>
          <div class="text-gray-500 text-sm">This folder doesn't contain any supported image files.</div>
        </div>
      </div>
    {:else}
      <div class="p-6">
        <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 xl:grid-cols-8 gap-4">
          {#each images as image, index}
            <div
              class="relative group cursor-pointer"
              on:click={() => toggleSelect(image, index)}
            >
              <!-- Selection indicator -->
              <div
                class="absolute top-2 right-2 w-6 h-6 rounded-full border-2 z-10 transition-all duration-200
                  {selectedIndices.has(index)
                    ? 'bg-blue-500 border-blue-500'
                    : 'bg-transparent border-white/50 group-hover:border-white'
                  }"
              >
                {#if selectedIndices.has(index)}
                  <div class="text-white text-sm font-bold flex items-center justify-center h-full">✓</div>
                {/if}
              </div>

              <!-- Image thumbnail -->
              <div
                class="aspect-square rounded-lg overflow-hidden border-2 transition-all duration-200
                  {selectedIndices.has(index)
                    ? 'border-blue-500 shadow-lg shadow-blue-500/25'
                    : 'border-transparent group-hover:border-gray-500'
                  }"
              >
                <img
                  src={api.getThumbnailUrl(image.path, 200)}
                  alt={image.name}
                  class="w-full h-full object-cover"
                  loading="lazy"
                />
              </div>

              <!-- Image info -->
              <div class="mt-2 text-xs text-gray-400 truncate">
                {image.name}
              </div>
            </div>
          {/each}
        </div>
      </div>
    {/if}
  </div>

  <!-- Footer with keyboard hints -->
  <div class="p-4 border-t border-gray-700 bg-gray-800">
    <div class="text-xs text-gray-400 space-x-4">
      <span><kbd class="kbd">Click</kbd> Select/Deselect</span>
      <span><kbd class="kbd">A</kbd> Select All</span>
      <span><kbd class="kbd">Enter</kbd> Compare</span>
      <span><kbd class="kbd">Esc</kbd> Back</span>
      <span><kbd class="kbd">?</kbd> Help</span>
    </div>
  </div>
</div>

<style>
  .kbd {
    @apply bg-gray-700 px-2 py-1 rounded text-xs font-mono border border-gray-600;
  }
</style>