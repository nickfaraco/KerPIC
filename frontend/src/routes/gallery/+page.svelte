<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { api } from '$lib/utils/api.js';
  import { selectedImages, selectedPhotos, markedForDeletion } from '$lib/stores/app.js';
  import PhotoGrid from '$lib/components/PhotoGrid.svelte';
  import PhotoViewer from '$lib/components/PhotoViewer.svelte';

  let loading = true;
  let error = null;
  let dashboardData = null;
  let recentPhotos = [];
  let albums = [];
  let stats = {};
  
  // Photo viewer state
  let viewerVisible = false;
  let currentPhotoIndex = 0;

  onMount(async () => {
    await loadDashboard();
  });

  async function loadDashboard() {
    loading = true;
    error = null;
    
    try {
      const dashboardData = await api.getDashboard();
      
      stats = dashboardData.stats;
      recentPhotos = dashboardData.recentPhotos || [];
      albums = dashboardData.recentAlbums || [];
      
      // Update the markedForDeletion store with photos from database
      const markedPaths = new Set();
      if (recentPhotos && Array.isArray(recentPhotos)) {
        recentPhotos.forEach(photo => {
          if (photo.markedForDeletion) {
            markedPaths.add(photo.path);
          }
        });
      }
      markedForDeletion.set(markedPaths);
      
      loading = false;
    } catch (err) {
      error = err.message;
      loading = false;
      console.error('Failed to load dashboard:', err);
    }
  }

  function handlePhotoSelect(event) {
    const photos = event.detail;
    // TODO: Implement photo selection for comparison
    console.log('Selected photos:', photos);
  }

  function handlePhotoView(event) {
    const { photo, index } = event.detail;
    currentPhotoIndex = index;
    viewerVisible = true;
  }

  function handleViewerClose() {
    viewerVisible = false;
  }

  function handleCompareSelected(event) {
    console.log('handleCompareSelected called, event.detail:', event.detail);
    console.log('Current selectedPhotos store:', $selectedPhotos);
    
    // Get the actual selected photos from the store
    const selectedPaths = Array.from($selectedPhotos);
    const selectedPhotoObjects = recentPhotos.filter(p => selectedPaths.includes(p.path));
    
    console.log('Selected photo objects for compare:', selectedPhotoObjects);
    
    if (selectedPhotoObjects && selectedPhotoObjects.length >= 2) {
      // Set the selected images in the store for the compare view
      selectedImages.set(selectedPhotoObjects);
      console.log('Set selectedImages store to:', selectedPhotoObjects);
      // Navigate to compare view using client-side routing
      goto('/gallery/compare');
    } else {
      console.log('Not enough photos selected for comparison:', selectedPhotoObjects.length);
    }
  }

  function handleAddToAlbum(event) {
    const photos = event.detail;
    console.log('Add to album:', photos);
    // TODO: Show album selection dialog
    const albumName = prompt('Enter album name (or select existing):');
    if (albumName) {
      // TODO: Implement add to album functionality
      console.log(`Adding ${photos.length} photos to album: ${albumName}`);
    }
  }

  async function handleDeleteMarked() {
    const markedPaths = Array.from($markedForDeletion);
    if (markedPaths.length === 0) {
      alert('No photos are marked for deletion.');
      return;
    }

    const confirmed = confirm(
      `Are you sure you want to permanently delete ${markedPaths.length} marked photos? This action cannot be undone.`
    );
    
    if (confirmed) {
      try {
        loading = true;
        await api.deleteMarkedPhotos();
        
        // Clear the marked photos from the store
        markedForDeletion.set(new Set());
        
        // Reload the dashboard to refresh the photo list
        await loadDashboard();
        
        alert(`Successfully deleted ${markedPaths.length} photos.`);
      } catch (error) {
        console.error('Failed to delete marked photos:', error);
        alert(`Error deleting photos: ${error.message}`);
        loading = false;
      }
    }
  }

  function handleNewAlbum() {
    // Navigate to album creation or show modal
    goto('/gallery/albums');
  }

  function handleImportPhotos() {
    // For now, just show an alert about file system access
    alert('To import photos, copy them to your mounted photos directory. The application will automatically detect them.');
  }
</script>

<svelte:head>
  <title>Gallery - KerPIC</title>
</svelte:head>

<div class="h-full flex flex-col">
  {#if loading}
    <div class="flex-1 flex items-center justify-center">
      <div class="text-center">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto"></div>
        <div class="mt-4 text-gray-400">Loading your photos...</div>
      </div>
    </div>
  {:else if error}
    <div class="flex-1 flex items-center justify-center">
      <div class="text-center">
        <div class="text-red-400 mb-2 text-xl">⚠ Error loading gallery</div>
        <div class="text-gray-400 mb-4">{error}</div>
        <button 
          class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
          on:click={loadDashboard}
        >
          Retry
        </button>
      </div>
    </div>
  {:else}
    <!-- Dashboard content -->
    <div class="flex-1 overflow-auto">
      <!-- Stats bar -->
      <div class="border-b border-opacity-20 px-6 py-6" 
           style="background-color: var(--bg-secondary); border-color: var(--color-dark-gray);">
        <div class="flex items-center space-x-12">
          <div class="text-center">
            <div class="text-2xl font-bold" style="color: var(--text-primary);">{stats.totalPhotos}</div>
            <div class="text-xs" style="color: var(--text-secondary);">Photos</div>
          </div>
          <div class="text-center">
            <div class="text-2xl font-bold" style="color: var(--text-primary);">{stats.totalAlbums}</div>
            <div class="text-xs" style="color: var(--text-secondary);">Albums</div>
          </div>
          <div class="text-center">
            <div class="text-2xl font-bold" style="color: var(--text-primary);">{stats.totalSize}</div>
            <div class="text-xs" style="color: var(--text-secondary);">Storage</div>
          </div>
        </div>
      </div>

      <!-- Main content -->
      <div class="p-6 space-y-8">
        <!-- Recent photos section -->
        <section>
          <div class="flex items-center justify-between mb-4">
            <h2 class="text-xl font-semibold" style="color: var(--text-primary);">Recent Photos</h2>
          </div>
          
          {#if recentPhotos.length > 0}
            <PhotoGrid 
              photos={recentPhotos} 
              columns={10}
              on:select={handlePhotoSelect}
              on:view={handlePhotoView}
              on:compare={handleCompareSelected}
              on:addToAlbum={handleAddToAlbum}
              on:deleteMarked={handleDeleteMarked}
            />
          {:else}
            <div class="text-center py-12">
              <div class="text-6xl mb-4">📷</div>
              <div class="text-xl mb-2" style="color: var(--text-secondary);">No photos found</div>
              <div class="text-sm" style="color: var(--text-secondary); opacity: 0.7;">Import some photos to get started</div>
            </div>
          {/if}
        </section>
      </div>
    </div>
  {/if}
</div>

<!-- Photo Viewer -->
<PhotoViewer 
  photos={recentPhotos} 
  bind:currentIndex={currentPhotoIndex}
  bind:visible={viewerVisible}
  on:close={handleViewerClose}
/>