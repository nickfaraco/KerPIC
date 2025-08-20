<script>
  import { onMount } from 'svelte';
  import { api } from '$lib/utils/api.js';
  import { currentView, selectedFolder } from '$lib/stores/app.js';
  
  let folders = [];
  let loading = true;
  let error = null;
  let selectedIndex = 0;

  onMount(async () => {
    try {
      folders = await api.getFolders();
      loading = false;
    } catch (err) {
      error = err.message;
      loading = false;
    }
  });

  async function selectFolder(folder, index) {
    selectedIndex = index;
    selectedFolder.set(folder);
    currentView.set('images');
  }

  function handleKeydown(event) {
    if (folders.length === 0) return;
    
    switch (event.key) {
      case 'ArrowUp':
        event.preventDefault();
        selectedIndex = Math.max(0, selectedIndex - 1);
        break;
      case 'ArrowDown':
        event.preventDefault();
        selectedIndex = Math.min(folders.length - 1, selectedIndex + 1);
        break;
      case 'Enter':
        event.preventDefault();
        selectFolder(folders[selectedIndex], selectedIndex);
        break;
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="h-full flex flex-col">
  <!-- Header -->
  <div class="p-6 border-b border-gray-700">
    <h2 class="text-xl font-semibold text-white mb-2">Select a Folder</h2>
    <p class="text-gray-400 text-sm">Choose a folder containing the images you want to organize.</p>
  </div>

  <!-- Content -->
  <div class="flex-1 p-6 overflow-auto">
    {#if loading}
      <div class="flex items-center justify-center h-64">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
        <span class="ml-3 text-gray-400">Loading folders...</span>
      </div>
    {:else if error}
      <div class="flex items-center justify-center h-64">
        <div class="text-center">
          <div class="text-red-400 mb-2">⚠ Error loading folders</div>
          <div class="text-gray-400 text-sm">{error}</div>
          <button 
            class="btn-primary mt-4"
            on:click={() => window.location.reload()}
          >
            Retry
          </button>
        </div>
      </div>
    {:else if folders.length === 0}
      <div class="flex items-center justify-center h-64">
        <div class="text-center">
          <div class="text-gray-400 mb-2">📁 No folders found</div>
          <div class="text-gray-500 text-sm">Make sure your photos directory is mounted and contains folders.</div>
        </div>
      </div>
    {:else}
      <div class="space-y-2">
        {#each folders as folder, index}
          <button
            class="w-full p-4 rounded-lg border-2 transition-all duration-200 text-left
              {index === selectedIndex 
                ? 'border-blue-500 bg-blue-500/10 text-white' 
                : 'border-gray-700 bg-gray-800 hover:border-gray-600 hover:bg-gray-750 text-gray-300'
              }"
            on:click={() => selectFolder(folder, index)}
          >
            <div class="flex items-center">
              <div class="text-2xl mr-3">📁</div>
              <div>
                <div class="font-medium">{folder.name}</div>
                <div class="text-sm opacity-70">Click to browse images</div>
              </div>
            </div>
          </button>
        {/each}
      </div>
    {/if}
  </div>

  <!-- Footer with keyboard hints -->
  <div class="p-4 border-t border-gray-700 bg-gray-800">
    <div class="text-xs text-gray-400 space-x-4">
      <span><kbd class="kbd">↑↓</kbd> Navigate</span>
      <span><kbd class="kbd">Enter</kbd> Select</span>
      <span><kbd class="kbd">?</kbd> Help</span>
    </div>
  </div>
</div>

<style>
  .kbd {
    @apply bg-gray-700 px-2 py-1 rounded text-xs font-mono border border-gray-600;
  }
</style>