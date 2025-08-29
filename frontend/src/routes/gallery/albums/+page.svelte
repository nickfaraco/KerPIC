<script>
  import { onMount } from 'svelte';
  import { api } from '$lib/utils/api.js';

  let albums = [];
  let loading = true;
  let error = null;

  onMount(async () => {
    await loadAlbums();
  });

  async function loadAlbums() {
    loading = true;
    error = null;
    
    try {
      const response = await api.getAlbums();
      albums = response.albums || [];
      loading = false;
    } catch (err) {
      error = err.message;
      loading = false;
      console.error('Failed to load albums:', err);
    }
  }

  function handleCreateAlbum() {
    // Simple prompt for album name for now
    const albumName = prompt('Enter album name:');
    if (albumName && albumName.trim()) {
      createAlbum(albumName.trim());
    }
  }

  async function createAlbum(name) {
    try {
      await api.createAlbum({
        name: name,
        description: '',
        photoPaths: [],
        tags: []
      });
      
      // Reload albums
      await loadAlbums();
    } catch (err) {
      alert('Failed to create album: ' + err.message);
    }
  }
</script>

<svelte:head>
  <title>Albums - KerPIC</title>
</svelte:head>

<div class="h-full flex flex-col">
  <div class="p-6">
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-white">Albums</h1>
      <button 
        class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
        on:click={handleCreateAlbum}
      >
        + Create Album
      </button>
    </div>

    {#if loading}
      <div class="flex items-center justify-center h-64">
        <div class="text-center">
          <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto"></div>
          <div class="mt-4 text-gray-400">Loading albums...</div>
        </div>
      </div>
    {:else}
      <!-- Albums grid -->
      <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-6 gap-6">
        <!-- Create new album card -->
        <button 
          class="aspect-square bg-gray-800 rounded-lg border-2 border-dashed border-gray-600 hover:border-gray-500 flex items-center justify-center transition-colors group"
          on:click={handleCreateAlbum}
        >
          <div class="text-center">
            <div class="text-4xl mb-2 group-hover:scale-110 transition-transform">📁</div>
            <div class="text-sm text-gray-400 group-hover:text-gray-300">Create Album</div>
          </div>
        </button>

        <!-- Existing albums -->
        {#each albums as album}
          <div class="aspect-square bg-gray-800 rounded-lg overflow-hidden border-2 border-gray-700 hover:border-gray-600 transition-colors">
            <div class="h-full flex flex-col">
              <!-- Album cover (placeholder for now) -->
              <div class="flex-1 bg-gradient-to-br from-gray-700 to-gray-800 flex items-center justify-center">
                <div class="text-4xl">📁</div>
              </div>
              <!-- Album info -->
              <div class="p-3 bg-gray-800">
                <div class="text-sm font-medium text-white truncate">{album.name}</div>
                <div class="text-xs text-gray-400">{album.photoCount || 0} photos</div>
              </div>
            </div>
          </div>
        {/each}
      </div>

      <!-- Empty state -->
      {#if albums.length === 0}
        <div class="flex items-center justify-center h-64 mt-8">
          <div class="text-center">
            <div class="text-6xl mb-4">📁</div>
            <div class="text-xl text-gray-400 mb-2">No albums yet</div>
            <div class="text-sm text-gray-500 mb-6">Create your first album to organize your photos</div>
            <button class="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors">
              Create Your First Album
            </button>
          </div>
        </div>
      {/if}
    {/if}
  </div>
</div>