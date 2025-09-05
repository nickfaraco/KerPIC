<script>
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { showHelp, thumbnailSize } from '$lib/stores/app.js';
  import { api } from '$lib/utils/api.js';
  import HelpModal from '$lib/components/HelpModal.svelte';
  import { onMount } from 'svelte';

  let albums = [];
  let sidebarCollapsed = false;

  onMount(async () => {
    // Load albums for sidebar
    try {
      const response = await api.getAlbums();
      albums = response.albums || [];
    } catch (err) {
      console.error('Failed to load albums:', err);
      albums = [];
    }
  });

  function handleGlobalKeydown(event) {
    if (event.key === '?') {
      event.preventDefault();
      showHelp.update(show => !show);
    }
  }

  function toggleSidebar() {
    sidebarCollapsed = !sidebarCollapsed;
  }
</script>

<svelte:head>
  <title>KerPIC - Photo Gallery</title>
</svelte:head>

<svelte:window on:keydown={handleGlobalKeydown} />

<div class="min-h-screen flex" style="background-color: var(--bg-primary); color: var(--text-primary);">
  <!-- Collapsible Side Panel -->
  <aside class="transition-all duration-300 ease-in-out flex-shrink-0 border-r border-opacity-20"
         style="background-color: var(--bg-secondary); border-color: var(--color-dark-gray);"
         class:w-64={!sidebarCollapsed}
         class:w-12={sidebarCollapsed}>

    <!-- Sidebar toggle -->
    <button
      class="w-full p-3 text-left transition-colors duration-200"
      style="color: var(--text-secondary);"
      on:click={toggleSidebar}
    >
      {sidebarCollapsed ? '📁' : '📁 Albums'}
    </button>

    {#if !sidebarCollapsed}
      <!-- Albums list -->
      <div class="px-3 pb-3">
        <div class="space-y-1">
          <button
            class="w-full px-3 py-2 text-left text-sm rounded transition-colors duration-200 hover:bg-opacity-50"
            style="color: var(--text-primary); background-color: transparent;"
            on:click={() => goto('/gallery')}
            class:font-medium={$page.url.pathname === '/gallery'}
          >
            All Photos
          </button>

          {#each albums as album}
            <button
              class="w-full px-3 py-2 text-left text-sm rounded transition-colors duration-200 hover:bg-opacity-50 truncate"
              style="color: var(--text-secondary); background-color: transparent;"
              on:click={() => goto(`/gallery/albums/${album.id}`)}
              title={album.name}
            >
              {album.name}
            </button>
          {/each}

          <!-- Create new album -->
          <button
            class="w-full px-3 py-2 text-left text-sm rounded transition-colors duration-200 border-dashed border opacity-50 hover:opacity-100"
            style="color: var(--text-secondary); border-color: var(--color-dark-gray);"
            on:click={() => goto('/gallery/albums')}
          >
            + New Album
          </button>
        </div>

        <!-- Thumbnail Size Slider -->
        <div class="px-3 py-3 border-t border-opacity-20" style="border-color: var(--color-dark-gray);">
          <div class="space-y-2">
            <label class="text-xs font-medium" style="color: var(--text-secondary);">
              Thumbnail Size
            </label>
            <div class="space-y-1">
              <input
                type="range"
                min="80"
                max="300"
                bind:value={$thumbnailSize}
                class="w-full h-3 rounded-lg appearance-none cursor-pointer"
                style="background: #6a6a6a; border: 1px solid #888;"
              />
              <div class="flex justify-between text-xs" style="color: var(--text-secondary);">
                <span>Small</span>
                <span>Large</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    {/if}
  </aside>

  <!-- Main content area -->
  <main class="flex-1 relative">
    <!-- Help hint in top-right corner -->
    <div class="absolute top-4 right-4 z-10 text-xs opacity-50" style="color: var(--text-secondary);">
      Press ? for help
    </div>

    <slot />
  </main>
</div>

<!-- Global components -->
<HelpModal />

<style>
  /* Custom range slider styling */
  input[type="range"] {
    -webkit-appearance: none;
    background: transparent;
  }

  input[type="range"]::-webkit-slider-track {
    width: 100%;
    height: 12px;
    cursor: pointer;
    border-radius: 6px;
    background: #6a6a6a;
    border: 1px solid #888;
    box-shadow: inset 0 1px 3px rgba(0,0,0,0.3);
  }

  input[type="range"]::-webkit-slider-thumb {
    border: none;
    height: 20px;
    width: 20px;
    border-radius: 50%;
    background: var(--accent);
    cursor: pointer;
    -webkit-appearance: none;
    margin-top: -5px;
    box-shadow: 0 2px 4px rgba(0,0,0,0.4);
  }

  input[type="range"]::-webkit-slider-thumb:hover {
    background: #F5CB5C;
    transform: scale(1.1);
  }

  input[type="range"]::-moz-range-track {
    width: 100%;
    height: 12px;
    cursor: pointer;
    border-radius: 6px;
    background: #6a6a6a;
    border: 1px solid #888;
    box-shadow: inset 0 1px 3px rgba(0,0,0,0.3);
  }

  input[type="range"]::-moz-range-thumb {
    border: none;
    height: 16px;
    width: 16px;
    border-radius: 50%;
    background: var(--accent);
    cursor: pointer;
    box-shadow: 0 2px 4px rgba(0,0,0,0.4);
  }
</style>