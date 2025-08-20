<script>
  import { onMount } from 'svelte';
  import { api } from '$lib/utils/api.js';
  import { currentView, selectedImages } from '$lib/stores/app.js';
  
  let currentPath = '';
  let loading = true;
  let error = null;
  let folderContents = null;
  let selectedImageIndices = [];
  let breadcrumbParts = [];
  
  $: folders = folderContents ? folderContents.subfolders || [] : [];
  $: images = folderContents ? folderContents.images || [] : [];
  $: selectedCount = selectedImageIndices.length;
  $: canCompare = selectedCount >= 2;

  onMount(async () => {
    await loadFolder('');
  });

  async function loadFolder(path) {
    loading = true;
    error = null;
    selectedImageIndices = []; // Clear selection when changing folders
    
    try {
      if (path === '') {
        // Load root folder
        const foldersResponse = await api.getFolders();
        folderContents = foldersResponse[0]; // Root contents
        currentPath = '';
        breadcrumbParts = [];
      } else {
        // Load subfolder
        folderContents = await api.getFolderContents(path);
        currentPath = path;
        breadcrumbParts = path.split('/').filter(p => p);
      }
      loading = false;
    } catch (err) {
      error = err.message;
      loading = false;
    }
  }

  function navigateToFolder(folderName) {
    const newPath = currentPath ? `${currentPath}/${folderName}` : folderName;
    loadFolder(newPath);
  }

  function navigateToBreadcrumb(index) {
    if (index === -1) {
      // Go to root
      loadFolder('');
    } else {
      // Go to specific path level
      const newPath = breadcrumbParts.slice(0, index + 1).join('/');
      loadFolder(newPath);
    }
  }

  function toggleImageSelection(index) {
    if (selectedImageIndices.includes(index)) {
      selectedImageIndices = selectedImageIndices.filter(i => i !== index);
    } else {
      selectedImageIndices = [...selectedImageIndices, index];
    }
  }

  function selectAllImages() {
    if (selectedImageIndices.length === images.length) {
      selectedImageIndices = [];
    } else {
      selectedImageIndices = images.map((_, i) => i);
    }
  }

  function startComparison() {
    if (!canCompare) return;
    
    const selected = selectedImageIndices.map(index => images[index]);
    selectedImages.set(selected);
    currentView.set('compare');
  }

  function handleKeydown(event) {
    switch (event.key) {
      case 'a':
      case 'A':
        if (images.length > 0) {
          event.preventDefault();
          selectAllImages();
        }
        break;
      case 'Enter':
        if (canCompare) {
          event.preventDefault();
          startComparison();
        }
        break;
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="h-full flex flex-col">
  <!-- Header -->
  {#if breadcrumbParts.length > 0}
    <div class="p-6 border-b border-gray-700">
      <!-- Breadcrumb navigation -->
      <nav class="text-sm text-gray-400">
        <button 
          class="hover:text-white"
          on:click={() => navigateToBreadcrumb(-1)}
        >
          📁 Root
        </button>
        {#each breadcrumbParts as part, index}
          <span class="mx-2">/</span>
          {#if index === breadcrumbParts.length - 1}
            <span class="text-white">{part}</span>
          {:else}
            <button 
              class="hover:text-white"
              on:click={() => navigateToBreadcrumb(index)}
            >
              {part}
            </button>
          {/if}
        {/each}
      </nav>
    </div>
  {/if}

  <!-- Content -->
  <div class="flex-1 p-6 overflow-auto">
    {#if loading}
      <div class="flex items-center justify-center h-64">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
        <span class="ml-3 text-gray-400">Loading...</span>
      </div>
    {:else if error}
      <div class="flex items-center justify-center h-64">
        <div class="text-center">
          <div class="text-red-400 mb-2">⚠ Error loading folder</div>
          <div class="text-gray-400 text-sm">{error}</div>
          <button class="btn-primary mt-4" on:click={() => loadFolder(currentPath)}>
            Retry
          </button>
        </div>
      </div>
    {:else}
      <div class="space-y-8">
        <!-- Subfolders -->
        {#if folders.length > 0}
          <div>
            <h3 class="text-lg font-medium text-white mb-3">📁 Folders</h3>
            <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-4">
              {#each folders as folder}
                <button
                  class="p-4 rounded-lg border-2 border-gray-700 bg-gray-800 hover:border-gray-600 hover:bg-gray-750 text-gray-300 transition-all duration-200"
                  on:click={() => navigateToFolder(folder.name)}
                >
                  <div class="text-center">
                    <div class="text-3xl mb-2">📁</div>
                    <div class="text-sm font-medium truncate">{folder.name}</div>
                  </div>
                </button>
              {/each}
            </div>
          </div>
        {/if}

        <!-- Images -->
        {#if images.length > 0}
          <div>
            <div class="flex items-center justify-between mb-3">
              <h3 class="text-lg font-medium text-white">
                📷 Images ({images.length})
                {#if selectedCount > 0}
                  <span class="text-blue-400">• {selectedCount} selected</span>
                {/if}
              </h3>
              <button
                class="text-sm text-gray-400 hover:text-white"
                on:click={selectAllImages}
              >
                {selectedImageIndices.length === images.length ? 'Deselect All' : 'Select All'}
              </button>
            </div>

            <div class="grid grid-cols-4 sm:grid-cols-6 md:grid-cols-8 lg:grid-cols-10 gap-3">
              {#each images as image, index}
                <button
                  class="aspect-square rounded overflow-hidden border-2 transition-all duration-200 hover:border-blue-500 focus:border-blue-500 focus:outline-none relative"
                  class:border-blue-500={selectedImageIndices.includes(index)}
                  class:border-gray-600={!selectedImageIndices.includes(index)}
                  on:click={() => toggleImageSelection(index)}
                  title={image.name}
                >
                  <!-- Selection indicator -->
                  {#if selectedImageIndices.includes(index)}
                    <div class="absolute top-1 right-1 w-5 h-5 bg-blue-500 rounded-full flex items-center justify-center text-xs text-white font-bold z-10">
                      ✓
                    </div>
                  {/if}
                  <img
                    src={api.getThumbnailUrl(image.path, 150)}
                    alt={image.name}
                    class="w-full h-full object-cover"
                    loading="lazy"
                  />
                </button>
              {/each}
            </div>
          </div>
        {/if}

        <!-- Empty state -->
        {#if !loading && folders.length === 0 && images.length === 0}
          <div class="flex items-center justify-center h-64">
            <div class="text-center">
              <div class="text-gray-400 mb-2">📂 Empty folder</div>
              <div class="text-gray-500 text-sm">This folder contains no images or subfolders.</div>
            </div>
          </div>
        {/if}
      </div>
    {/if}
  </div>

  <!-- Footer -->
  <div class="px-6 py-2 border-t border-gray-700 bg-gray-800">
    <div class="flex items-center justify-between">
      <div class="text-xs text-gray-400 space-x-4">
        <span><kbd class="kbd">Click</kbd> Select images</span>
        <span><kbd class="kbd">A</kbd> Select/deselect all</span>
        <span><kbd class="kbd">Enter</kbd> Compare</span>
        <span><kbd class="kbd">?</kbd> Help</span>
      </div>
      
      <div class="flex items-center gap-4" class:invisible={selectedCount === 0}>
        <span class="text-sm text-gray-400">
          {selectedCount} image{selectedCount !== 1 ? 's' : ''} selected
        </span>
        <button
          class="btn-primary"
          disabled={!canCompare}
          on:click={startComparison}
        >
          Compare {selectedCount} Images
        </button>
      </div>
    </div>
  </div>
</div>

<style>
  .kbd {
    @apply bg-gray-700 px-2 py-1 rounded text-xs font-mono border border-gray-600;
  }
</style>