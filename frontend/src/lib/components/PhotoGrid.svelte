<script>
  import { createEventDispatcher } from 'svelte';
  import { api } from '$lib/utils/api.js';
  import { 
    selectedPhotos, 
    markedForDeletion,
    toggleSelection,
    addMultipleToSelection,
    addToSelection,
    removeFromSelection,
    markForDeletion,
    unmarkForDeletion,
    undoLastAction,
    clearSelection,
    thumbnailSize
  } from '$lib/stores/app.js';

  export let photos = [];
  export let columns = 8;
  export let selected = [];

  const dispatch = createEventDispatcher();

  let gridContainer;
  let containerWidth = 0;
  let isDragging = false;
  let dragStartX, dragStartY;
  let dragSelection = new Set();
  let selectionBox = null;

  // Dynamic grid calculation
  let calculatedColumns = 4;
  let gridStyle = '';
  let debouncedThumbnailSize = $thumbnailSize;
  let debounceTimeout;
  
  // Debounce thumbnail size changes to prevent too many API requests
  $: if ($thumbnailSize !== debouncedThumbnailSize) {
    clearTimeout(debounceTimeout);
    debounceTimeout = setTimeout(() => {
      debouncedThumbnailSize = $thumbnailSize;
    }, 100);
  }
  
  $: if (containerWidth && $thumbnailSize) {
    const gap = 8; // Tailwind gap-2 = 8px
    const padding = 0; // No padding on the grid container itself
    const availableWidth = containerWidth - padding;
    const itemWidth = $thumbnailSize + gap;
    calculatedColumns = Math.max(1, Math.floor(availableWidth / itemWidth));
    
    // Create CSS custom properties for the grid
    gridStyle = `
      display: grid;
      grid-template-columns: repeat(${calculatedColumns}, ${$thumbnailSize}px);
      gap: ${gap}px;
      justify-content: start;
    `;
  }


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

  function handlePhotoClick(photo, index, event) {
    // Prevent action if this was part of a drag operation
    if (isDragging) return;
    
    console.log('Photo clicked:', photo.name, 'Marked for deletion:', isMarkedForDeletion(photo));
    
    // If photo is marked for deletion, clicking it unmarks it
    if (isMarkedForDeletion(photo)) {
      console.log('Unmarking photo for deletion:', photo.path);
      unmarkPhotoForDeletion(photo.path);
      return;
    }
    
    // Single click toggles selection
    console.log('Toggling selection for:', photo.path);
    toggleSelection(photo.path);
  }
  
  function handlePhotoDoubleClick(photo, index) {
    // Double click opens viewer
    console.log('Double-clicking to open viewer for:', photo.name);
    viewPhoto(photo, index);
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
    console.log('PhotoGrid keydown:', event.key);
    
    // Global shortcuts
    if (event.key === 'x' || event.key === 'X') {
      event.preventDefault();
      dispatch('deleteMarked');
      return;
    }
    
    if (event.key === 'a' || event.key === 'A') {
      if (photos.length > 0) {
        event.preventDefault();
        selectAll();
      }
      return;
    }
    
    // Selection shortcuts (only work when photos are selected)
    if ($selectedPhotos.size > 0) {
      if (event.key === 'd' || event.key === 'D') {
        event.preventDefault();
        markSelectedForDeletion();
      } else if (event.key === 'r' || event.key === 'R') {
        event.preventDefault();
        unmarkSelectedForDeletion();
      } else if (event.key === 'c' || event.key === 'C') {
        event.preventDefault();
        compareSelected();
      } else if (event.key === 'Escape') {
        event.preventDefault();
        clearSelection();
      }
    }
    
    if (event.key === 'u' || event.key === 'U') {
      event.preventDefault();
      undoLastAction().catch(error => {
        console.error('Failed to undo action:', error);
      });
    }
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

  async function unmarkPhotoForDeletion(photoPath) {
    try {
      await unmarkForDeletion([photoPath]);
      console.log('Successfully unmarked photo for deletion:', photoPath);
    } catch (error) {
      console.error('Failed to unmark photo for deletion:', error);
      // Could show user-friendly error message here
    }
  }

  async function unmarkSelectedForDeletion() {
    const selectedPaths = Array.from($selectedPhotos);
    const markedSelectedPaths = selectedPaths.filter(path => $markedForDeletion.has(path));
    
    if (markedSelectedPaths.length > 0) {
      try {
        await unmarkForDeletion(markedSelectedPaths);
        console.log('Successfully unmarked selected photos for deletion:', markedSelectedPaths);
        // Keep photos selected so user can see which ones were unmarked
      } catch (error) {
        console.error('Failed to unmark selected photos for deletion:', error);
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

  // Drag selection functions
  function handleMouseDown(event) {
    if (event.button !== 0) return; // Only left mouse button
    
    isDragging = false;
    dragStartX = event.clientX;
    dragStartY = event.clientY;
    dragSelection = new Set();
    
    // Add mouse move and up listeners to document
    document.addEventListener('mousemove', handleMouseMove);
    document.addEventListener('mouseup', handleMouseUp);
  }
  
  function handleMouseMove(event) {
    const deltaX = Math.abs(event.clientX - dragStartX);
    const deltaY = Math.abs(event.clientY - dragStartY);
    
    // Start dragging if moved more than 5 pixels
    if (!isDragging && (deltaX > 5 || deltaY > 5)) {
      isDragging = true;
      createSelectionBox();
    }
    
    if (isDragging && selectionBox) {
      updateSelectionBox(event);
      updateDragSelection(event);
    }
  }
  
  function handleMouseUp(event) {
    document.removeEventListener('mousemove', handleMouseMove);
    document.removeEventListener('mouseup', handleMouseUp);
    
    if (isDragging) {
      // Apply drag selection based on modifier keys
      if (dragSelection.size > 0) {
        const draggedPaths = Array.from(dragSelection);
        
        if (event.ctrlKey || event.metaKey) {
          // Ctrl/Cmd + drag = toggle selection for dragged photos
          draggedPaths.forEach(path => toggleSelection(path));
        } else if (event.shiftKey) {
          // Shift + drag = deselect dragged photos
          draggedPaths.forEach(path => removeFromSelection(path));
        } else {
          // Normal drag = select dragged photos
          addMultipleToSelection(draggedPaths);
        }
      }
      removeSelectionBox();
    }
    
    // Reset drag state after a short delay to prevent click events
    setTimeout(() => {
      isDragging = false;
    }, 10);
  }
  
  function createSelectionBox() {
    selectionBox = document.createElement('div');
    selectionBox.style.position = 'fixed';
    selectionBox.style.border = '2px solid #F5CB5C';
    selectionBox.style.backgroundColor = 'rgba(245, 203, 92, 0.2)';
    selectionBox.style.pointerEvents = 'none';
    selectionBox.style.zIndex = '1000';
    document.body.appendChild(selectionBox);
  }
  
  function updateSelectionBox(event) {
    if (!selectionBox) return;
    
    const left = Math.min(dragStartX, event.clientX);
    const top = Math.min(dragStartY, event.clientY);
    const width = Math.abs(event.clientX - dragStartX);
    const height = Math.abs(event.clientY - dragStartY);
    
    selectionBox.style.left = left + 'px';
    selectionBox.style.top = top + 'px';
    selectionBox.style.width = width + 'px';
    selectionBox.style.height = height + 'px';
  }
  
  function updateDragSelection(event) {
    if (!gridContainer) return;
    
    const selectionRect = {
      left: Math.min(dragStartX, event.clientX),
      top: Math.min(dragStartY, event.clientY),
      right: Math.max(dragStartX, event.clientX),
      bottom: Math.max(dragStartX, event.clientY)
    };
    
    dragSelection.clear();
    
    // Check each photo element for intersection with selection box
    const photoButtons = gridContainer.querySelectorAll('button');
    photoButtons.forEach((button, index) => {
      const rect = button.getBoundingClientRect();
      
      if (rect.left < selectionRect.right &&
          rect.right > selectionRect.left &&
          rect.top < selectionRect.bottom &&
          rect.bottom > selectionRect.top) {
        // Photo intersects with selection box
        if (index < photos.length) {
          dragSelection.add(photos[index].path);
        }
      }
    });
  }
  
  function removeSelectionBox() {
    if (selectionBox) {
      document.body.removeChild(selectionBox);
      selectionBox = null;
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
    <!-- Stats -->
    <div class="flex items-center justify-between mb-4 px-4">
      <div class="text-sm" style="color: var(--text-secondary);">
        <span>{photos.length} photos</span>
        {#if $selectedPhotos.size > 0}
          • <span style="color: var(--accent);">{$selectedPhotos.size} selected</span>
        {/if}
        {#if $markedForDeletion.size > 0}
          • <span style="color: var(--danger);">{$markedForDeletion.size} marked for deletion</span>
        {/if}
      </div>
      
    </div>
  {/if}

  <!-- Photo grid -->
  <div 
    bind:this={gridContainer}
    bind:clientWidth={containerWidth}
    style={gridStyle}
    on:mousedown={handleMouseDown}
    role="grid"
  >
    {#each photos as photo, index}
      <button
        class="aspect-square rounded overflow-hidden transition-all duration-200 focus:outline-none relative group"
        style="width: {$thumbnailSize}px; height: {$thumbnailSize}px; border: {$markedForDeletion.has(photo.path) ? '4px solid #ef4444' : $selectedPhotos.has(photo.path) ? '4px solid #F5CB5C' : '1px solid var(--color-dark-gray)'}; box-shadow: {$selectedPhotos.has(photo.path) && !$markedForDeletion.has(photo.path) ? '0 0 0 2px rgba(245, 203, 92, 0.6)' : 'none'};"
        on:click={(event) => handlePhotoClick(photo, index, event)}
        on:dblclick={() => handlePhotoDoubleClick(photo, index)}
        title={photo.name}
      >

        <!-- Photo -->
        <img
          src={api.getThumbnailUrl(photo.path, debouncedThumbnailSize)}
          alt={photo.name}
          key={`${photo.path}-${debouncedThumbnailSize}`}
          class="w-full h-full object-contain transition-transform duration-200 group-hover:scale-105"
          style="opacity: {$markedForDeletion.has(photo.path) ? '0.3' : '1'};
                 filter: {$markedForDeletion.has(photo.path) ? 'saturate(0.2) grayscale(0.8)' : 'none'};"
          loading="lazy"
          on:error={(event) => {
            // Force reload on error by changing src slightly
            if (event.target.src.includes('?')) return;
            event.target.src = api.getThumbnailUrl(photo.path, debouncedThumbnailSize) + '?retry=' + Date.now();
          }}
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

  {#if $selectedPhotos.size > 0 || $markedForDeletion.size > 0}
    <!-- Keyboard shortcuts -->
    <div class="mt-6 text-center px-4">
      <div class="text-xs space-x-4" style="color: var(--text-secondary);">
        <span><kbd class="kbd">A</kbd> Select all</span>
        <span><kbd class="kbd">Esc</kbd> Clear selection</span>
        <span><kbd class="kbd">Shift+Drag</kbd> Deselect</span>
        <span><kbd class="kbd">Ctrl+Drag</kbd> Toggle</span>
        {#if $selectedPhotos.size > 0}
          <span><kbd class="kbd">D</kbd> Mark for deletion</span>
          <span><kbd class="kbd">R</kbd> Restore from deletion</span>
          {#if $selectedPhotos.size >= 2}
            <span><kbd class="kbd">C</kbd> Compare selected</span>
          {/if}
        {/if}
        <span><kbd class="kbd">U</kbd> Undo</span>
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