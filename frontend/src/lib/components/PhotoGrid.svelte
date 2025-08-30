<script>
  import { createEventDispatcher } from 'svelte';
  import { api } from '$lib/utils/api.js';
  import { 
    selectionMode, 
    selectedPhotos, 
    markedForDeletion,
    toggleSelection,
    markForDeletion,
    undoLastAction,
    clearSelection
  } from '$lib/stores/app.js';

  export let photos = [];
  export let columns = 8;
  export let selected = [];

  const dispatch = createEventDispatcher();

  let gridContainer;

  // Reactive grid columns based on screen size and prop
  $: gridCols = `grid-cols-${Math.min(columns, 12)}`;
  $: responsiveGridCols = `grid-cols-4 sm:grid-cols-6 md:grid-cols-${Math.min(columns, 8)} lg:grid-cols-${columns}`;


  function viewPhoto(photo, index) {
    // Dispatch photo view event
    dispatch('view', { photo, index });
  }

  function isSelected(photo) {
    return $selectedPhotos.has(photo.path);
  }

  function isMarkedForDeletion(photo) {
    return $markedForDeletion.has(photo.path);
  }

  function handlePhotoClick(photo, index) {
    console.log('Photo clicked:', photo.name, 'Selection mode:', $selectionMode);
    if ($selectionMode) {
      // In selection mode, clicking toggles selection
      console.log('Toggling selection for:', photo.path);
      toggleSelection(photo.path);
    } else {
      // In normal mode, clicking opens viewer
      console.log('Opening viewer for:', photo.name);
      viewPhoto(photo, index);
    }
  }

  function selectAll() {
    selectedPhotos.update(currentSet => {
      if (currentSet.size === photos.length) {
        return new Set();
      } else {
        return new Set(photos.map(p => p.path));
      }
    });
  }

  function handleKeydown(event) {
    console.log('PhotoGrid keydown:', event.key, 'Selection mode:', $selectionMode);
    
    // Global shortcuts
    if (event.key === 's' || event.key === 'S') {
      event.preventDefault();
      console.log('S key pressed, toggling selection mode');
      toggleSelectionMode();
      return;
    }

    if (event.key === 'x' || event.key === 'X') {
      event.preventDefault();
      dispatch('deleteMarked');
      return;
    }

    // Selection mode shortcuts
    if ($selectionMode) {
      if (event.key === 'd' || event.key === 'D') {
        event.preventDefault();
        markSelectedForDeletion();
      } else if (event.key === 'u' || event.key === 'U') {
        event.preventDefault();
        undoLastAction().catch(error => {
          console.error('Failed to undo action:', error);
        });
      } else if (event.key === 'c' || event.key === 'C') {
        event.preventDefault();
        compareSelected();
      } else if (event.key === 'a' || event.key === 'A') {
        event.preventDefault();
        addToAlbumPrompt();
      }
    } else {
      // Normal mode shortcuts
      if (event.key === 'a' || event.key === 'A') {
        if (photos.length > 0) {
          event.preventDefault();
          selectAll();
        }
      }
    }
  }

  function toggleSelectionMode() {
    selectionMode.update(mode => {
      const newMode = !mode;
      console.log(`Selection mode toggled from ${mode} to ${newMode}`);
      if (!newMode) {
        // Exiting selection mode - clear selections
        clearSelection();
      }
      return newMode;
    });
  }

  async function markSelectedForDeletion() {
    const selectedPaths = Array.from($selectedPhotos);
    if (selectedPaths.length > 0) {
      try {
        await markForDeletion(selectedPaths);
        // Remove marked photos from selection since they're no longer viable
        clearSelection();
      } catch (error) {
        console.error('Failed to mark photos for deletion:', error);
        // Could show user-friendly error message here
      }
    }
  }

  function compareSelected() {
    const selectedPaths = Array.from($selectedPhotos);
    const selectedPhotoObjects = photos.filter(p => selectedPaths.includes(p.path));
    if (selectedPhotoObjects.length >= 2) {
      dispatch('compare', selectedPhotoObjects);
    }
  }

  function addToAlbumPrompt() {
    const selectedPaths = Array.from($selectedPhotos);
    const selectedPhotoObjects = photos.filter(p => selectedPaths.includes(p.path) && !$markedForDeletion.has(p.path));
    if (selectedPhotoObjects.length > 0) {
      dispatch('addToAlbum', selectedPhotoObjects);
    }
  }

  // Update selected when prop changes (legacy support)
  $: if (selected && selected.length > 0) {
    const paths = selected.map(p => p.path || p);
    selectedPhotos.set(new Set(paths));
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="photo-grid-container">
  {#if photos.length > 0}
    <!-- Mode indicator and stats -->
    <div class="flex items-center justify-between mb-4 px-4">
      <div class="text-sm" style="color: var(--text-secondary);">
        {#if $selectionMode}
          <span style="color: var(--accent);">Selection Mode</span>
          {#if $selectedPhotos.size > 0}
            • <span style="color: var(--text-primary);">{$selectedPhotos.size} selected</span>
          {/if}
          {#if $markedForDeletion.size > 0}
            • <span style="color: var(--danger);">{$markedForDeletion.size} marked for deletion</span>
          {/if}
          <!-- Debug info -->
          <span style="color: #ff0000; font-size: 10px;"> [DEBUG: SM={$selectionMode}, S={$selectedPhotos.size}, D={$markedForDeletion.size}]</span>
        {:else}
          <span>{photos.length} photos</span>
          <!-- Debug info -->
          <span style="color: #ff0000; font-size: 10px;"> [DEBUG: SM={$selectionMode}]</span>
        {/if}
      </div>
      
      {#if $selectionMode}
        <div class="text-xs" style="color: var(--text-secondary);">
          Press S to exit selection mode
        </div>
      {:else}
        <div class="text-xs" style="color: var(--text-secondary);">
          Press S to enter selection mode
        </div>
      {/if}
    </div>
  {/if}

  <!-- Photo grid -->
  <div 
    bind:this={gridContainer}
    class="grid gap-2 {responsiveGridCols}"
  >
    {#each photos as photo, index}
      <button
        class="aspect-square rounded overflow-hidden transition-all duration-200 focus:outline-none relative group"
        on:click={() => handlePhotoClick(photo, index)}
        title={photo.name}
        style="border: {$selectedPhotos.has(photo.path) ? '4px solid #F5CB5C' : $markedForDeletion.has(photo.path) ? '4px solid #ef4444' : '1px solid var(--color-dark-gray)'}; 
               box-shadow: {$selectedPhotos.has(photo.path) ? '0 0 0 2px rgba(245, 203, 92, 0.6)' : 'none'};"
      >

        <!-- Photo -->
        <img
          src={api.getThumbnailUrl(photo.path, 200)}
          alt={photo.name}
          class="w-full h-full object-cover transition-transform duration-200 group-hover:scale-105"
          style="opacity: {$markedForDeletion.has(photo.path) ? '0.3' : '1'};
                 filter: {$markedForDeletion.has(photo.path) ? 'saturate(0.2) grayscale(0.8)' : 'none'};"
          loading="lazy"
        />

        <!-- Hover overlay with photo info -->
        <div class="absolute inset-0 bg-black bg-opacity-0 group-hover:bg-opacity-50 transition-all duration-200 flex items-end p-2">
          <div class="text-white text-xs opacity-0 group-hover:opacity-100 transition-opacity duration-200 truncate w-full">
            {photo.name}
          </div>
        </div>
      </button>
    {/each}
  </div>

  {#if photos.length === 0}
    <div class="flex items-center justify-center h-64">
      <div class="text-center">
        <div class="text-gray-400 mb-2">📷 No photos to display</div>
        <div class="text-gray-500 text-sm">Photos will appear here when available.</div>
      </div>
    </div>
  {/if}

  {#if $selectionMode && ($selectedPhotos.size > 0 || $markedForDeletion.size > 0)}
    <!-- Selection mode keyboard shortcuts -->
    <div class="mt-6 text-center px-4">
      <div class="text-xs space-x-4" style="color: var(--text-secondary);">
        <span><kbd class="kbd">Click</kbd> Toggle selection</span>
        <span><kbd class="kbd">D</kbd> Mark for deletion</span>
        <span><kbd class="kbd">U</kbd> Undo</span>
        {#if $selectedPhotos.size >= 2}
          <span><kbd class="kbd">C</kbd> Compare selected</span>
        {/if}
        {#if $selectedPhotos.size > 0}
          <span><kbd class="kbd">A</kbd> Add to album</span>
        {/if}
        {#if $markedForDeletion.size > 0}
          <span><kbd class="kbd">X</kbd> Delete marked</span>
        {/if}
      </div>
    </div>
  {/if}
</div>

<style>
  .photo-grid-container {
    width: 100%;
  }
  
  @media (max-width: 640px) {
    .photo-grid-container :global(.grid) {
      grid-template-columns: repeat(3, minmax(0, 1fr));
      gap: 0.25rem;
    }
  }
  
  @media (min-width: 641px) and (max-width: 768px) {
    .photo-grid-container :global(.grid) {
      grid-template-columns: repeat(5, minmax(0, 1fr));
      gap: 0.5rem;
    }
  }
</style>